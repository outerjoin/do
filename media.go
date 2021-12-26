package do

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rightjoin/fig"
	"github.com/rightjoin/rutl/conv"
)

func NewMedia(webContext interface{}, referenceStructOrStr interface{}, formInput string) (*Media, error) {

	c, isEcho := webContext.(echo.Context)
	if !isEcho {
		panic("must be echooo")
	}

	// reference (what does this media belong to?)
	// is stored in DB and is also part of media's
	// prefix src
	prefix := ""
	if s, ok := referenceStructOrStr.(string); ok {
		prefix = s
	} else if TypeComposedOf(referenceStructOrStr, MongoEntity{}) {
		prefix = MongoCollectionName(referenceStructOrStr)
	} else {
		prefix = TypeOf(referenceStructOrStr).Name()
	}
	prefix = strings.ReplaceAll(prefix, "_", "-")
	prefix = conv.CaseURL(prefix)
	if prefix != "" {
		prefix += "/" + conv.CaseURL(formInput)
	}

	// load fileHeader
	fileHeader, err := c.FormFile(formInput)
	if err != nil {
		return nil, err
	}
	if fileHeader == nil {
		return nil, nil
	}

	// load file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var b bytes.Buffer
	b.ReadFrom(file)

	id := NewAlhpaNum(9)
	mime := http.DetectContentType(b.Bytes())
	var p_w, p_h *int
	if strings.HasPrefix(mime, "image/") {
		img, _, err := image.DecodeConfig(bytes.NewReader(b.Bytes()))
		if err == nil {
			p_w = &img.Width
			p_h = &img.Height
		}
	}
	path := CleanFilePath(fmt.Sprintf("/%s/%s/%s/%s/%s",
		prefix,
		id[0:3], id[3:6], id[6:],
		CleanFileName(fileHeader.Filename)))

	return &Media{
		ID:     id,
		Src:    path,
		Mime:   mime,
		Size:   uint(fileHeader.Size),
		Width:  p_w,
		Height: p_h,

		// DB fields
		Buffer:    b,
		Reference: prefix,
		Extn:      filepath.Ext(path[1:]),
	}, nil
}

type Media struct {
	ID     string `bson:"_id" json:"id" insert:"no" auto:"alphanum(9)"`
	Src    string `bson:"src" json:"src"`
	Mime   string `bson:"mime" json:"mime"`
	Size   uint   `bson:"size" json:"size"`
	Width  *int   `bson:"width" json:"width"`
	Height *int   `bson:"height" json:"height"`

	// Other fields, that are saved to DB
	// but not loaded to simplify Media object
	Buffer    bytes.Buffer `bson:"-" json:"-"`
	Reference string       `bson:"reference" json:"-"` // what object does this file contextually belong to
	Extn      string       `bson:"extn" json:"-"`      // file extension
	Timed     `bson:"inline" json:"-"`
}

func (m Media) Serialize() error {

	folder := fig.StringOr("./media", "media.folder")
	path := CleanFilePath(folder + "/" + m.Src)

	// created nested folder path
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}

	// touch file
	out, err := os.Create(path)
	if err != nil {
		return err
	}

	// write file
	_, err = io.Copy(out, bytes.NewReader(m.Buffer.Bytes()))
	if err != nil {
		return err
	}

	return nil
}
