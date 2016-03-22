package base

import (
	"bytes"
	"reflect"
	//	"io/ioutil"
	"encoding/csv"
	//	"github.com/djimenez/iconv-go"
	//.	"logger"
)

// 生成表头
func FetchCSVFieldName(data interface{}) []string {

	typ := reflect.TypeOf(data)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() == reflect.Slice {

		typ = typ.Elem()

		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
	}

	var ret = make([]string, 0, typ.NumField())

	for i := 0; i < typ.NumField(); i++ {
		structField := typ.Field(i)

		fieldName := structField.Tag.Get("excel")

		if fieldName == "-" {
			continue
		}

		if fieldName == "" {
			fieldName = structField.Name
		}

		ret = append(ret, fieldName)
	}

	//	DEBUG( "CSV FIELD: %s %v", typ.Name(), ret )
	return ret
}

type IExcelString interface {
	ExcelString() string
}

var void_param = make([]reflect.Value, 0)

// 生成一行内容
func FetchFieldValue(data interface{}) []string {

	typ := reflect.TypeOf(data)
	val := reflect.ValueOf(data)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	var ret = make([]string, 0, typ.NumField())

	for i := 0; i < typ.NumField(); i++ {
		field := val.Field(i)

		if typ.Field(i).Tag.Get("excel") == "-" {
			continue
		}

		var value interface{}

		switch field.Kind() {
		case reflect.Ptr:
			if field.IsNil() {
				value = ""
				break
			}
			field = field.Elem()
			fallthrough
		default:
			if field.Type().Implements(reflect.TypeOf((*IExcelString)(nil)).Elem()) {
				method, ok := field.Type().MethodByName("ExcelString")
				Assert(ok, "No Method ExcelString()")
				rets := method.Func.Call([]reflect.Value{field})
				value = rets[0].Interface()
			} else {
				value = field.Interface()
			}
		}

		ret = append(ret, Sprintf("%v", value))
	}

	return ret
}

func BuildCSVEx(meta interface{}, geter func() []interface{}) []byte {

	buf := bytes.NewBufferString("")
	//	buf.WriteString( "\xEF\xBB\xBF" )	// UTF-BOM

	w := csv.NewWriter(buf)
	w.Write(FetchCSVFieldName(meta))

	for {

		data := geter()

		if 0 == len(data) {
			break
		}

		typ := reflect.TypeOf(data)
		val := reflect.ValueOf(data)

		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if typ.Kind() == reflect.Slice {

			for i := 0; i < val.Len(); i++ {
				w.Write(FetchFieldValue(val.Index(i).Interface()))
			}

		} else {
			w.Write(FetchFieldValue(data))
		}
	}

	w.Flush()
	return buf.Bytes()

	/*
		reader, err := iconv.NewReader( buf, "utf-8", "gbk")

		if err != nil {
			LOG_ERROR( "转换编码错误: %s", err.Error() )
			return buf.Bytes()
		}

		bytes, err := ioutil.ReadAll( reader )

		if err != nil {
			LOG_ERROR( "读取失败: %s", err.Error() )
			return buf.Bytes()
		}

		return bytes
	*/
}

// 生成csv文件
func BuildCSV(data interface{}) []byte {

	buf := bytes.NewBufferString("")
	//	buf.WriteString( "\xEF\xBB\xBF" )	// UTF-BOM

	w := csv.NewWriter(buf)
	w.Write(FetchCSVFieldName(data))

	typ := reflect.TypeOf(data)
	val := reflect.ValueOf(data)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if typ.Kind() == reflect.Slice {

		for i := 0; i < val.Len(); i++ {
			w.Write(FetchFieldValue(val.Index(i).Interface()))
		}

	} else {
		w.Write(FetchFieldValue(data))
	}

	w.Flush()
	return buf.Bytes()
	/*
		reader, err := iconv.NewReader( buf, "utf-8", "gbk")

		if err != nil {
			LOG_ERROR( "转换编码错误: %s", err.Error() )
			return buf.Bytes()
		}

		bytes, err := ioutil.ReadAll( reader )

		if err != nil {
			LOG_ERROR( "读取失败: %s", err.Error() )
			return buf.Bytes()
		}

		return bytes
	*/
}

// 转换编码 "Hello", "UTF-8", "GBK"
/*func Convert( str, from, to string ) string {

	output, err := iconv.ConvertString( str, from, to )

	if err != nil {
		LOG_ERROR( "编码转换失败: %s", err.Error() )
		return str
	}

	return output
}*/

///////////////////////////////////////////////////////////////////////////////////////////////

type ExcelBuilder struct {
	buffer *bytes.Buffer
	writer *csv.Writer
}

func NewExcelBuilderWith(fields []string) *ExcelBuilder {

	buffer := bytes.NewBufferString("")
	writer := csv.NewWriter(buffer)

	// excel field name.
	writer.Write(fields)

	return &ExcelBuilder{
		buffer: buffer,
		writer: writer,
	}
}

func NewExcelBuilder(meta interface{}) *ExcelBuilder {
	return NewExcelBuilderWith(FetchCSVFieldName(meta))
}

func (this *ExcelBuilder) Append(data interface{}) {

	typ := reflect.TypeOf(data)
	val := reflect.ValueOf(data)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if typ.Kind() == reflect.Slice {

		for i := 0; i < val.Len(); i++ {
			this.writer.Write(FetchFieldValue(val.Index(i).Interface()))
		}

	} else {
		this.writer.Write(FetchFieldValue(data))
	}
}

func (this *ExcelBuilder) AppendValues(values []string) {
	this.writer.Write(values)
}

func (this *ExcelBuilder) Save(filename string, encoding ...string) error {

	this.writer.Flush()

	/*
		if len(encoding) > 0 && encoding[0] != "utf-8" {

			reader, err := iconv.NewReader( this.buffer, "utf-8", encoding[0])

			if err != nil {
				ERROR( "转换编码错误: %s", err.Error() )
				return err
			}

			return WriteStream( filename, reader )
		}
	*/
	return WriteStream(filename, this.buffer)
}
