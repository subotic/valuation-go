package main

import (
	_ "embed"
	"fmt"
	"github.com/subotic/valuation-go/documents/calculator"
	"github.com/subotic/valuation-go/types"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/schema"
	datastar "github.com/starfederation/datastar/sdk/go"
)

//go:embed hello-world.html
var helloWorldHTML []byte

func main() {
	r := chi.NewRouter()

	const message = "Hello, world!"
	type Store struct {
		Delay time.Duration `json:"delay"` // delay in milliseconds between each character of the message.
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write(helloWorldHTML)
		if err != nil {
			return
		}
	})

	r.Get("/hello-world", func(w http.ResponseWriter, r *http.Request) {
		store := &Store{}
		if err := datastar.ReadSignals(r, store); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		sse := datastar.NewSSE(w, r)

		for i := 0; i < len(message); i++ {
			err := sse.MergeFragments(`<div id="message">` + message[:i+1] + `</div>`)
			if err != nil {
				return
			}
			time.Sleep(store.Delay * time.Millisecond)
		}
	})

	r.Get("/calculator", func(w http.ResponseWriter, r *http.Request) {
		err := calculator.IndexPage().Render(r.Context(), w)
		if err != nil {
			return
		}
	})
	r.Get("/calculator/styles.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		_, err := w.Write(calculator.StyleCSS)
		if err != nil {
			return
		}
	})
	r.Get("/calculator/calculate", handleCalculate)

	log.Println("Go to: http://localhost:8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		return
	}
}

var decoder = schema.NewDecoder()

func handleCalculate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid query params", http.StatusBadRequest)
		return
	}

	var form types.DcfForm
	if err := decoder.Decode(&form, r.Form); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	ctx := computeDcfTable(
		form.FCF,
		form.Growth,
		form.Discount,
		form.Terminal,
		form.Years,
	)
	view := types.ToDcfTableView(ctx)

	// Render HTML snippet from Templ (ctx must match template expectations)
	tableFragment := calculator.TableFragment(view) // returns string

	sse := datastar.NewSSE(w, r)
	err := sse.MergeFragments(fmt.Sprintf("<div id='intrinsic_value' class='value-display'>Intrinsic Value: %f </div>", ctx.TotalIntrinsicValue))
	if err != nil {
		http.Error(w, "Failed to send SSE message", http.StatusInternalServerError)
		return
	}
	err = sse.MergeFragmentTempl(tableFragment)
	if err != nil {
		http.Error(w, "Failed to send SSE message", http.StatusInternalServerError)
		return
	}
}

func computeDcfTable(fcf, growth, discount, terminal float64, years uint) types.DcfTableContext {
	rows := make([]types.CashFlowRow, 0)
	total := 0.0

	for year := uint(1); year <= years; year++ {
		power := float64(year)
		projectedFcf := fcf * math.Pow(1+growth, power)
		discounted := projectedFcf / math.Pow(1+discount, power)

		rows = append(rows, types.CashFlowRow{
			Year:       fmt.Sprintf("%d", year),
			FCF:        projectedFcf,
			Discounted: discounted,
		})

		total += discounted
	}

	lastFcf := fcf * math.Pow(1+growth, float64(years))
	terminalValue := lastFcf * (1 + terminal) / (discount - terminal)
	discountedTerminal := terminalValue / math.Pow(1+discount, float64(years))

	rows = append(rows, types.CashFlowRow{
		Year:       "Terminal",
		FCF:        roundToTwoDecimals(terminalValue),
		Discounted: roundToTwoDecimals(discountedTerminal),
	})

	total += discountedTerminal

	return types.DcfTableContext{
		Rows:                rows,
		TotalIntrinsicValue: roundToTwoDecimals(total),
	}
}

func roundToTwoDecimals(value float64) float64 {
	return math.Round(value*100) / 100
}
