package house

import (
	"encoding/xml"
	"log"
	"os"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pavlik/fias_xml2postgresql/helpers"
)

func ExportBulk(w *sync.WaitGroup, c chan string, db *sqlx.DB, format *string, logger *log.Logger) {

	defer w.Done()

	helpers.DropAndCreateTable(schema, tableName, db)

	var format2 string
	format2 = *format
	fileName, err2 := helpers.SearchFile(tableName+"_", format2)
	if err2 != nil {
		logger.Fatalln("Error searching file:", err2)
	}

	pathToFile := format2 + "/" + fileName

	xmlFile, err := os.Open(pathToFile)
	if err != nil {
		logger.Fatalln("Error opening file:", err)
	}

	defer xmlFile.Close()

	decoder := xml.NewDecoder(xmlFile)
	var inElement string
	total := 0
	i := 0

	txn, err := db.Begin()
	if err != nil {
		logger.Fatalln(err)
	}

	query := pq.CopyIn(tableName,
		"house_guid",
		"postal_code",
		"ifns_fl",
		"terr_ifns_fl",
		"ifns_ul",
		"terr_ifns_ul",
		"okato",
		"oktmo",
		"update_date",
		"house_num",
		"est_status",
		"build_num",
		"struc_num",
		"str_status",
		"house_id",
		"ao_guid",
		"start_date",
		"end_date",
		"stat_status",
		"norm_doc",
		"counter")

	stmt, err := txn.Prepare(query)
	if err != nil {
		logger.Fatalln(err)
	}

	for {
		if i == 100000 {
			i = 0

			_, err = stmt.Exec()
			if err != nil {
				logger.Fatalln(err)
			}

			err = stmt.Close()
			if err != nil {
				logger.Fatalln(err)
			}

			err = txn.Commit()
			if err != nil {
				logger.Fatalln(err)
			}

			//c <- helpers.PrintRowsAffected(elementName, total)

			txn, err = db.Begin()
			if err != nil {
				logger.Fatalln(err)
			}

			stmt, err = txn.Prepare(query)
			if err != nil {
				logger.Fatalln(err)
			}
		}
		// Read tokens from the XML document in a stream.
		t, _ := decoder.Token()

		// Если достигли конца xml-файла
		if t == nil {
			if i > 0 {
				_, err = stmt.Exec()
				if err != nil {
					logger.Fatalln(err)
				}

				err = stmt.Close()
				if err != nil {
					logger.Fatalln(err)
				}

				err = txn.Commit()
				if err != nil {
					logger.Fatalln(err)
				}
			}

			//c <- helpers.PrintRowsAffected(elementName, total)

			break
		}
		// Inspect the type of the token just read.
		switch se := t.(type) {
		case xml.StartElement:
			// If we just read a StartElement token
			inElement = se.Name.Local

			if inElement == elementName {
				var item XmlObject

				// decode a whole chunk of following XML into the
				// variable item which is a ActualStatus (se above)
				err = decoder.DecodeElement(&item, &se)
				if err != nil {
					logger.Fatalln("Error in decode element:", err)
				}

				_, err = stmt.Exec(item.HOUSEGUID,
					item.POSTALCODE,
					item.IFNSFL,
					item.TERRIFNSFL,
					item.IFNSUL,
					item.TERRIFNSUL,
					item.OKATO,
					item.OKTMO,
					item.UPDATEDATE,
					item.HOUSENUM,
					item.ESTSTATUS,
					item.BUILDNUM,
					item.STRUCNUM,
					item.STRSTATUS,
					item.HOUSEID,
					item.AOGUID,
					item.STARTDATE,
					item.ENDDATE,
					item.STATSTATUS,
					item.NORMDOC,
					item.COUNTER)

				if err != nil {
					logger.Fatalln(err)
				}
				total++
				i++
				c <- helpers.PrintRowsAffected(elementName, total)
			}
		default:
		}

	}
}
