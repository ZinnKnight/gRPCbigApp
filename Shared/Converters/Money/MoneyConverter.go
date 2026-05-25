package Money

import (
	"github.com/shopspring/decimal"
	decimalpb "google.golang.org/genproto/googleapis/type/decimal"
	moneypb "google.golang.org/genproto/googleapis/type/money"
)

// как я понял по доке и примерам что нашёл - мы тут описываем что именно и в какую валюту будем конвертить

const Currency = "USD"

const nanosPerUnit = 1_000_000_000

func MoneyToDec(m *moneypb.Money) decimal.Decimal {
	if m == nil {
		return decimal.Zero
	}
	units := decimal.NewFromInt(m.GetUnits())
	nanos := decimal.New(int64(m.GetNanos()), -9)
	return units.Add(nanos)
}

func DecToMoney(d decimal.Decimal, currency string) *moneypb.Money {
	if currency == "" {
		currency = Currency
	}

	units := d.IntPart()
	frac := d.Sub(decimal.NewFromInt(units))
	nanos := int32(frac.Mul(decimal.NewFromInt(nanosPerUnit)).IntPart())
	return &moneypb.Money{
		CurrencyCode: currency,
		Units:        units,
		Nanos:        nanos,
	}
}

// из decimal в decimalPB и обратно, что бы можно было типы расшифровывать

func DecimalPBToDecimal(d *decimalpb.Decimal) (decimal.Decimal, error) {
	if d == nil || d.GetValue() == "" {
		return decimal.Zero, nil
	}
	return decimal.NewFromString(d.GetValue())
}

func DecimalToDecimalPB(d decimal.Decimal) *decimalpb.Decimal {
	return &decimalpb.Decimal{
		Value: d.String(),
	}
}
