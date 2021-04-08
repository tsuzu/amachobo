package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/jszwec/csvutil"
	"github.com/ktr0731/go-fuzzyfinder"
)

type Receipt struct {
	Date                string `csv:"注文日"`
	OrderNo             string `csv:"注文番号"`
	Name                string `csv:"商品名"`
	Attribute           string `csv:"付帯情報"`
	Price               *int   `csv:"価格"`
	Count               *int   `csv:"個数"`
	PartialSum          *int   `csv:"商品小計"`
	Sum                 *int   `csv:"注文合計"`
	Address             string `csv:"お届け先"`
	Status              string `csv:"状態"`
	Payer               string `csv:"請求先"`
	BillingAmount       string `csv:"請求額"`
	BillingDate         string `csv:"クレカ請求日"`
	CreditBillingAmount *int   `csv:"クレカ請求額"`
	BillingKind         string `csv:"クレカ種類"`
	OrderURL            string `csv:"注文概要URL"`
	ReceiptURL          string `csv:"領収書URL"`
	ItemURL             string `csv:"商品URL"`
}

type Freee struct {
	InOutcome          string `csv:"収支区分"`
	ManagementNo       string `csv:"管理番号"`
	Date               string `csv:"発生日"`
	DueDate            string `csv:"決済期日"`
	ClientCode         string `csv:"取引先コード"`
	Client             string `csv:"取引先"`
	ItemAccount        string `csv:"勘定科目"`
	TaxSection         string `csv:"税区分"`
	Amount             string `csv:"金額"`
	TaxCalculationKind string `csv:"税計算区分"`
	Tax                string `csv:"税額"`
	Additional         string `csv:"備考"`
	ItemKind           string `csv:"品目"`
	Section            string `csv:"部門"`
	Note               string `csv:"メモタグ（複数指定可、カンマ区切り）"`
	Segment1           string `csv:"セグメント1"`
	Segment2           string `csv:"セグメント2"`
	Segment3           string `csv:"セグメント3"`
	BillingDate        string `csv:"決済日"`
	BillingAccount     string `csv:"決済口座"`
	BillingAmount      string `csv:"決済金額"`
}

func newFreeeFromReceipt(r *Receipt) *Freee {
	sum := fmt.Sprint((*r.Count) * (*r.Price))
	notes := []string{}
	itemAccount := "消耗品費"

	kindle := strings.HasPrefix(r.Attribute, "[Kindle 版]")

	if kindle {
		notes = append(notes, "Kindle")
		itemAccount = "新聞図書費"
	}

	fr := &Freee{
		InOutcome:          "支出",
		ManagementNo:       r.OrderNo,
		Date:               r.Date,
		Client:             "Amazon",
		ItemAccount:        itemAccount, // 消耗品費, 新聞図書費
		TaxSection:         "課対仕入10%",
		Amount:             sum,
		TaxCalculationKind: "内税",
		Tax:                "",
		Additional:         r.Name,
		ItemKind:           "Amazon",
		Section:            "",
		Note:               strings.Join(notes, ","),
		Segment1:           "",
		Segment2:           "",
		Segment3:           "",
		BillingDate:        r.Date,
		BillingAccount:     "事業主借",
		BillingAmount:      sum,
	}

	return fr
}

func newCSVReader(r io.Reader) *csv.Reader {
	br := bufio.NewReader(r)
	bs, err := br.Peek(3)
	if err != nil {
		return csv.NewReader(br)
	}
	if bs[0] == 0xEF && bs[1] == 0xBB && bs[2] == 0xBF {
		br.Discard(3)
	}
	return csv.NewReader(br)
}

func parseCSV(path string) ([]*Receipt, error) {
	fp, err := os.Open(path)

	if err != nil {
		return nil, err
	}
	defer fp.Close()

	reader := newCSVReader(fp)

	decoder, err := csvutil.NewDecoder(reader)

	if err != nil {
		return nil, err
	}

	rcpts := make([]*Receipt, 0)
	for {
		var rcpt Receipt
		if err := decoder.Decode(&rcpt); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		rcpts = append(rcpts, &rcpt)
	}

	return rcpts, nil
}

func exportCSV(writer io.Writer, rcpts []*Freee) error {
	csvWriter := csv.NewWriter(writer)
	encoder := csvutil.NewEncoder(csvWriter)

	for _, rcpt := range rcpts {
		if err := encoder.Encode(rcpt); err != nil {
			return err
		}
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return err
	}

	return nil
}

func numberOrEmpty(v *int) string {
	if v == nil {
		return ""
	}

	return fmt.Sprint(*v)
}

func selectReceipts(rcpts []*Receipt) ([]*Receipt, error) {
	selected, err := fuzzyfinder.FindMulti(rcpts, func(i int) string {
		return rcpts[i].Name
	}, fuzzyfinder.WithPreviewWindow(func(i, width, height int) string {
		rcpt := rcpts[i]
		return fmt.Sprintf("注文日: %v\n", rcpt.Date) +
			fmt.Sprintf("注文番号: %v\n", rcpt.OrderNo) +
			fmt.Sprintf("商品名: %v\n", rcpt.Name) +
			fmt.Sprintf("付帯情報: %v\n", rcpt.Attribute) +
			fmt.Sprintf("価格: %v\n", numberOrEmpty(rcpt.Price)) +
			fmt.Sprintf("個数: %v\n", numberOrEmpty(rcpt.Count)) +
			fmt.Sprintf("商品小計: %v\n", numberOrEmpty(rcpt.PartialSum)) +
			fmt.Sprintf("注文合計: %v\n", numberOrEmpty(rcpt.Sum)) +
			fmt.Sprintf("お届け先: %v\n", rcpt.Address) +
			fmt.Sprintf("状態: %v\n", rcpt.Status) +
			fmt.Sprintf("請求先: %v\n", rcpt.Payer) +
			fmt.Sprintf("請求額: %v\n", rcpt.BillingAmount) +
			fmt.Sprintf("クレカ請求日: %v\n", rcpt.BillingDate) +
			fmt.Sprintf("クレカ請求額: %v\n", numberOrEmpty(rcpt.CreditBillingAmount)) +
			fmt.Sprintf("クレカ種類: %v\n", rcpt.BillingKind) +
			fmt.Sprintf("注文概要URL: %v\n", rcpt.OrderURL) +
			fmt.Sprintf("領収書URL: %v\n", rcpt.ReceiptURL) +
			fmt.Sprintf("商品URL: %v\n", rcpt.ItemURL)
	}))

	if err != nil {
		return nil, err
	}

	res := make([]*Receipt, 0, len(selected))
	for i := range selected {
		res = append(res, rcpts[selected[i]])
	}

	return res, nil
}

func filterNoises(rcpts []*Receipt) []*Receipt {
	new := make([]*Receipt, 0, len(rcpts))

	for _, rcpt := range rcpts {
		if strings.HasPrefix(rcpt.Name, "（") && strings.HasSuffix(rcpt.Name, "）") {
			continue
		}

		new = append(new, rcpt)
	}

	return new
}

func load(p string) []*Receipt {
	selected, err := parseCSV(p)

	if err != nil {
		panic(err)
	}

	selected, err = selectReceipts(filterNoises(selected))

	if err != nil {
		panic(err)
	}

	return selected
}

func main() {
	selected := make([]*Receipt, 0)
	for _, p := range os.Args[1:] {
		selected = append(selected, load(p)...)
	}

	frs := make([]*Freee, 0, len(selected))
	for i := range selected {
		frs = append(frs, newFreeeFromReceipt(selected[i]))
	}

	fp, err := os.Create(time.Now().Format("freee-2006-01-02.csv"))

	if err != nil {
		panic(err)
	}
	defer fp.Close()

	if err := exportCSV(fp, frs); err != nil {
		panic(err)
	}
}
