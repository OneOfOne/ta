package csvticks

import (
	"compress/gzip"
	"errors"
	"io"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.oneofone.dev/ta"
)

func init() {
	const DefaultDateTimeFormat = `2006-01-02T15:04:05+0000`
	defaultTimeFmt.Store(DefaultDateTimeFormat)
}

var (
	ErrMissingMapping = errors.New("missing mapping")

	typeCache sync.Map
	timeType  = reflect.TypeOf(time.Time{})

	defaultTimeFmt atomic.Value
)

func GzipDecoder(r io.Reader) (io.Reader, error) {
	return gzip.NewReader(r)
}

// SetDefaultTimeFormat sets the default time format for DateTime parsing,
// and it returns the previous value.
// the default is `2006-01-02T15:04:05+0000`
func SetDefaultTimeFormat(f string) string {
	old, _ := defaultTimeFmt.Load().(string)
	defaultTimeFmt.Store(f)
	return old
}

// DateTime is a generic helper date/time type since every
// dataset uses a different format for their timestamp
type DateTime string

func (dt *DateTime) UnmarshalJSON(b []byte) error {
	*dt = DateTime(b)
	return nil
}

func (dt DateTime) MarshalJSON() ([]byte, error) {
	return []byte(dt), nil
}

// Time tries to convert the DateTime to time.Time
// rules:
// - if it's a number, it tries to parse it as nanoseconds, milliseconds or seconds
// - if it's a string it'll try to parse it with the default time fmt
// see `SetDefaultTimeFormat`
func (dt DateTime) Time() time.Time {
	const ms = int64(1e12)
	const ns = int64(1e15)
	if dt[0] == '"' {
		dtfmt, _ := defaultTimeFmt.Load().(string)
		t, _ := time.Parse(dtfmt, string(dt[1:len(dt)-1]))
		return t
	}
	n, _ := strconv.ParseInt(string(dt), 10, 64)

	if n > ns {
		return time.Unix(0, n)
	}
	if n > ms {
		n /= 1000
	}
	return time.Unix(n, 0)
}

// ConvertToTicks converts a slice of structs to a Ticks, it expects:
// Symbol, if available, to be string
// TS|Date|DateTime, if available, to be a string, time.Time or int64
// Open|High|Low|Closem, if available, to be float(32|64) or an alias for it.
// Volume to be int type.
func ConvertToTicks(v interface{}, nameMapping map[string]string) (out Ticks) {
	rv, flds, ok := getV(v, nameMapping)
	if !ok {
		panic("expected a slice of structs with at least one matching field")
	}
	out = make(Ticks, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		it := reflect.Indirect(rv.Index(i))
		var (
			t  Tick
			ok bool
			ev reflect.Value
		)

		if flds.TS != nil {
			if ev = reflect.Indirect(it.FieldByIndex(flds.TS)); !ev.IsValid() {
				continue
			}
			switch ev.Kind() {
			case reflect.Struct:
				if v, ok := ev.Interface().(time.Time); ok {
					t.TS = DateTime(strconv.FormatInt(v.UnixNano(), 10))
				}
				ok = true
			case reflect.Int64:
				t.TS = DateTime(strconv.FormatInt(ev.Int(), 10))
				ok = true
			case reflect.String:
				t.TS = DateTime(ev.String())
				ok = true
			}
		}

		if flds.Symbol != nil {
			if ev = reflect.Indirect(it.FieldByIndex(flds.Symbol)); !ev.IsValid() {
				continue
			}
			t.Symbol = ev.String()
			ok = true
		}

		if flds.Open != nil {
			if ev = reflect.Indirect(it.FieldByIndex(flds.Open)); !ev.IsValid() {
				continue
			}
			t.Open = ta.Decimal(ev.Float())
			ok = true
		}

		if flds.High != nil {
			if ev = reflect.Indirect(it.FieldByIndex(flds.High)); !ev.IsValid() {
				continue
			}
			t.High = ta.Decimal(ev.Float())
			ok = true
		}

		if flds.Low != nil {
			if ev = reflect.Indirect(it.FieldByIndex(flds.Low)); !ev.IsValid() {
				continue
			}
			t.Low = ta.Decimal(ev.Float())
			ok = true
		}

		if flds.Close != nil {
			if ev = reflect.Indirect(it.FieldByIndex(flds.Close)); !ev.IsValid() {
				continue
			}
			t.Close = ta.Decimal(ev.Float())
			ok = true
		}

		if flds.Volume != nil {
			if ev = reflect.Indirect(it.FieldByIndex(flds.Volume)); !ev.IsValid() {
				continue
			}
			t.Volume = ev.Int()
			ok = true
		}

		if ok {
			out = append(out, &t)
		}
	}
	return
}

type taFields struct {
	TS     []int
	Symbol []int
	Open   []int
	High   []int
	Low    []int
	Close  []int
	Volume []int
}

func getV(v interface{}, nameMapping map[string]string) (rv reflect.Value, fields *taFields, ok bool) {
	switch rv = reflect.ValueOf(v); rv.Kind() {
	case reflect.Array, reflect.Slice:
	default:
		return
	}

	et := rv.Type().Elem()
	if et.Kind() == reflect.Ptr {
		et = et.Elem()
	}
	if et.Kind() != reflect.Struct {
		return
	}

	if flds, _ := typeCache.Load(et); flds != nil {
		fields, _ = flds.(*taFields)
		ok = fields != nil
		println("hit cache")
		return
	}

	fields = &taFields{}
	for _, n := range []string{"TS", "Symbol", "Open", "High", "Low", "Close", "Volume"} {
		fn := n
		if nm, ok := nameMapping[n]; ok {
			fn = nm
		}
		if sf, _ := et.FieldByName(fn); sf.Index != nil {
			ok = true
			switch n {
			case "TS":
				fields.TS = sf.Index
			case "Symbol":
				fields.Symbol = sf.Index
			case "Open":
				fields.Open = sf.Index
			case "High":
				fields.High = sf.Index
			case "Low":
				fields.Low = sf.Index
			case "Close":
				fields.Close = sf.Index
			case "Volume":
				fields.Volume = sf.Index
			}
		}
	}

	if fields.TS == nil && nameMapping["TS"] == "" {
		for i := 0; i < et.NumField(); i++ {
			f := et.Field(i)
			n := strings.ToUpper(f.Name)
			k := f.Type.Kind()
			switch {
			case f.Type == timeType:
			case k == reflect.Int64 && strings.HasSuffix(n, "MS"):
			case k == reflect.String && (strings.Contains(n, "DATE") || strings.Contains(n, "TIME")):
			default:
				continue
			}
			fields.TS = f.Index
			ok = true
			break
		}
	}
	if !ok {
		fields = nil
	}
	typeCache.Store(et, fields)
	return
}
