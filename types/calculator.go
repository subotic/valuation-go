package types

import "fmt"

type DcfForm struct {
	FCF      float64 `query:"fcf"`
	Growth   float64 `query:"growth"`
	Discount float64 `query:"discount"`
	Terminal float64 `query:"terminal"`
	Years    uint    `query:"years"`
}

type CashFlowRow struct {
	Year       string
	FCF        float64
	Discounted float64
}

type DcfTableContext struct {
	Rows                []CashFlowRow
	TotalIntrinsicValue float64
}

type CashFlowRowView struct {
	Year       string
	FCF        string
	Discounted string
}

type DcfTableView struct {
	Rows                []CashFlowRowView
	TotalIntrinsicValue string
}

func ToDcfTableView(ctx DcfTableContext) DcfTableView {
	rows := make([]CashFlowRowView, 0, len(ctx.Rows))
	for _, r := range ctx.Rows {
		rows = append(rows, CashFlowRowView{
			Year:       r.Year,
			FCF:        fmt.Sprintf("%.2f", r.FCF),
			Discounted: fmt.Sprintf("%.2f", r.Discounted),
		})
	}

	return DcfTableView{
		Rows:                rows,
		TotalIntrinsicValue: fmt.Sprintf("%.2f", ctx.TotalIntrinsicValue),
	}
}
