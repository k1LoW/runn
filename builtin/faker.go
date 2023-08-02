package builtin

import (
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

type Faker struct {
	engine *gofakeit.Faker
}

func NewFaker() *Faker {
	return &Faker{
		engine: gofakeit.NewCrypto(),
	}
}

// https://github.com/brianvoe/gofakeit#person

func (f *Faker) Name() string      { return f.engine.Name() }
func (f *Faker) FirstName() string { return f.engine.FirstName() }
func (f *Faker) LastName() string  { return f.engine.LastName() }
func (f *Faker) Email() string     { return f.engine.Email() }

// https://github.com/brianvoe/gofakeit#auth

func (f *Faker) Username() string { return f.engine.Username() }
func (f *Faker) Password(lower bool, upper bool, numeric bool, special bool, space bool, num int) string {
	return f.engine.Password(lower, upper, numeric, special, space, num)
}

// https://github.com/brianvoe/gofakeit#misc

func (f *Faker) Bool() bool   { return f.engine.Bool() }
func (f *Faker) UUID() string { return f.engine.UUID() }

// https://github.com/brianvoe/gofakeit#colors

func (f *Faker) Color() string    { return f.engine.Color() }
func (f *Faker) HexColor() string { return f.engine.HexColor() }

// https://github.com/brianvoe/gofakeit#internet

func (f *Faker) URL() string         { return f.engine.URL() }
func (f *Faker) Domain() string      { return f.engine.DomainName() }
func (f *Faker) IPv4() string        { return f.engine.IPv4Address() }
func (f *Faker) IPv6() string        { return f.engine.IPv6Address() }
func (f *Faker) HTTPStatusCode() int { return f.engine.HTTPStatusCode() }
func (f *Faker) HTTPMethod() string  { return f.engine.HTTPMethod() }
func (f *Faker) HTTPVersion() string { return f.engine.HTTPVersion() }
func (f *Faker) UserAgent() string   { return f.engine.UserAgent() }

// https://github.com/brianvoe/gofakeit#datetime

func (f *Faker) Date() time.Time { return f.engine.Date() }
func (f *Faker) NanoSecond() int { return f.engine.NanoSecond() }
func (f *Faker) Second() int     { return f.engine.Second() }
func (f *Faker) Minute() int     { return f.engine.Minute() }
func (f *Faker) Hour() int       { return f.engine.Hour() }
func (f *Faker) Month() int      { return f.engine.Month() }
func (f *Faker) Day() int        { return f.engine.Day() }
func (f *Faker) Year() int       { return f.engine.Year() }

// https://github.com/brianvoe/gofakeit#emoji

func (f *Faker) Emoji() string { return f.engine.Emoji() }

// https://github.com/brianvoe/gofakeit#number

func (f *Faker) Int() int                            { return int(f.engine.Int64()) }
func (f *Faker) IntRange(min int, max int) int       { return f.engine.Number(min, max) }
func (f *Faker) Float() float64                      { return f.engine.Float64() }
func (f *Faker) FloatRange(min, max float64) float64 { return f.engine.Float64() }
func (f *Faker) RandomInt(i []int) int               { return f.engine.RandomInt(i) }

// https://github.com/brianvoe/gofakeit#string

func (f *Faker) Digit() string { return f.engine.Digit() }
func (f *Faker) DigitN(n int) string {
	if n < 0 {
		return ""
	}
	return f.engine.DigitN(uint(n))
}
func (f *Faker) Letter() string { return f.engine.Letter() }
func (f *Faker) LetterN(n int) string {
	if n < 0 {
		return ""
	}
	return f.engine.LetterN(uint(n))
}
func (f *Faker) Lexify(str string) string       { return f.engine.Lexify(str) }
func (f *Faker) Numerify(str string) string     { return f.engine.Numerify(str) }
func (f *Faker) RandomString(a []string) string { return f.engine.RandomString(a) }
