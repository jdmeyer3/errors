package main

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/jdmeyer3/errors"
	"github.com/jdmeyer3/errors/extchoozle"
	"google.golang.org/protobuf/proto"
	//"github.com/cockroachdb/errors"
)

type ChoErr []byte

type Foo struct {
	Name string `json:"name"`
	Err  ChoErr `json:"err"`
}

func (f ChoErr) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	zl := zlib.NewWriter(&buf)
	zl.Write(f)
	zl.Close()
	return []byte(`"` + base64.StdEncoding.EncodeToString(buf.Bytes()) + `"`), nil
}

func (f *ChoErr) UnmarshalJSON(data []byte) error {
	s := strings.Split(string(data), `"`)
	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(b)
	r, err := zlib.NewReader(buf)
	if err != nil {
		return err
	}
	bits, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	*f = bits
	return nil
}

func main() {
	ctx := context.Background()
	err := extchoozle.WrapWithChoozleError(errors.New("failed to make a friend"), extchoozle.ErrorProvider, "")

	err = errors.Wrap(err, "wrap")

	enc := errors.EncodeError(ctx, err)

	data, err := proto.Marshal(enc)
	//b, err := enc.Marshal()
	if err != nil {
		panic(err)
	}
	f := Foo{
		Name: "foo",
		Err:  data,
	}

	b, err := json.Marshal(f)
	if err != nil {
		panic(err)
	}

	f2 := Foo{}
	err = json.Unmarshal(b, &f2)
	if err != nil {
		panic(err)
	}

	enc2 := &errors.EncodedError{}
	encErr := proto.Unmarshal(f2.Err, enc2)

	if encErr != nil {
		panic(encErr)
	}

	dec := errors.DecodeError(ctx, enc2)

	fmt.Println(errors.UnwrapAll(dec))
	fmt.Println(errors.GetAllDetails(dec))
	st := errors.GetReportableStackTrace(dec)
	for i, fr := range st.Frames {
		if i == 0 {
			fmt.Println(fr.Function)
		} else {
			fmt.Printf("%#v\n", fr)
		}
	}
	fmt.Println(errors.ReportError(dec))
}
