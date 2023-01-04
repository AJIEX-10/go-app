package main

import (
	"database/sql"
	"encoding/xml"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/microcosm-cc/bluemonday"
)

func main() {
	type Staff struct {
		XMLName     xml.Name `xml:"item"`
		Id          string   `xml:"id,attr"`
		Name        string   `xml:"name"`
		Description string   `xml:"description"`
		Price       string   `xml:"price"`
	}

	type Offers struct {
		Staffs []Staff `xml:"item"`
	}

	type Times struct {
		Open  string `xml:"open"`
		Close string `xml:"close"`
	}

	type Magazine struct {
		ShopId   string   `xml:"id,attr"`
		ShopName string   `xml:"name"`
		Url      string   `xml:"url"`
		WorkTime []Times  `xml:"working_time"`
		Offerz   []Offers `xml:"offers"`
	}

	type Shop struct {
		XMLName xml.Name   `xml:"Shops"`
		Shops   []Magazine `xml:"shop"`
	}

	v := &Offers{}
	m := &Magazine{}
	sh := &Shop{}

	db, err := sql.Open("mysql", "root:@/shops")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM `магазины`")
	if err != nil {
		log.Fatal(err)
	}

	columns, err := rows.Columns()
	if err != nil {
		log.Fatal(err)
	}

	values := make([]sql.RawBytes, len(columns))

	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	var id_sh string
	var name_sh string
	var url string
	var times string
	var open string
	var close string

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			log.Fatal(err)
		}

		var value string
		for i, col := range values {
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}

			switch columns[i] {
			case "id":
				id_sh = value
			case "имя":
				name_sh = value
			case "url":
				url = value
			case "время работы":
				times = value
				t := strings.Split(times, "-")
				open = t[0]
				close = t[1]
			}
		}

		db, err := sql.Open("mysql", "root:@/shops")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		rows, err := db.Query("SELECT * FROM `товары`")
		if err != nil {
			log.Fatal(err)
		}

		columns, err := rows.Columns()
		if err != nil {
			log.Fatal(err)
		}

		values := make([]sql.RawBytes, len(columns))

		scanArgs := make([]interface{}, len(values))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		var id_val string
		var name string
		var descr string
		var price string
		var id_mag string

		p := bluemonday.StripTagsPolicy()

		for rows.Next() {
			err = rows.Scan(scanArgs...)
			if err != nil {
				log.Fatal(err)
			}

			var value string
			for i, col := range values {
				if col == nil {
					value = "NULL"
				} else {
					value = string(col)
				}

				html := p.Sanitize(value)

				switch columns[i] {
				case "id магазина":
					id_mag = value
				case "id":
					id_val = value
				case "название":
					name = value
				case "описание":
					descr = html
				case "цена":
					price = value
				}
			}

			if id_sh == id_mag {
				v.Staffs = append(v.Staffs, Staff{
					Id:          id_val,
					Name:        name,
					Description: descr,
					Price:       price,
				})
			}
		}

		m.WorkTime = append(m.WorkTime, Times{Open: open, Close: close})
		m.Offerz = append(m.Offerz, Offers{v.Staffs})
		sh.Shops = append(sh.Shops, Magazine{
			ShopId:   id_sh,
			ShopName: name_sh,
			Url:      url,
			WorkTime: m.WorkTime,
			Offerz:   m.Offerz,
		})

		if id_sh != id_mag {
			m.WorkTime = nil
			v.Staffs = nil
			m.Offerz = nil
		}

		if err = rows.Err(); err != nil {
			log.Fatal(err)
		}
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	data, err := xml.MarshalIndent(sh, " ", "    ")
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile("shops.xml", data, 0666)
	if err != nil {
		log.Fatal(err)
	}
}
