package elem

import (
	"database/sql"

	"github.com/Bitspark/slang/pkg/core"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"

	"reflect"
)

var databaseQueryCfg = &builtinConfig{
	safe: true,
	blueprint: core.Blueprint{
		Id: uuid.MustParse("ce3a3e0e-d579-4712-8573-713a645c2271"),
		Meta: core.BlueprintMetaDef{
			Name:             "DB query",
			ShortDescription: "queries an SQL query on a relational database and emits the result set",
			Icon:             "database",
			Tags:             []string{"database"},
			DocURL:           "https://bitspark.de/slang/docs/operator/db-execute",
		},
		ServiceDefs: map[string]*core.ServiceDef{
			core.MAIN_SERVICE: {
				In: core.TypeDef{
					Type: "map",
					Map: map[string]*core.TypeDef{
						"trigger": {
							Type: "trigger",
						},
						"{queryParams}": {
							Type: "primitive",
						},
					},
				},
				Out: core.TypeDef{
					Type: "stream",
					Stream: &core.TypeDef{
						Type: "map",
						Map: map[string]*core.TypeDef{
							"trigger": {
								Type: "trigger",
							},
							"{rowColumns}": {
								Type: "primitive",
							},
						},
					},
				},
			},
		},
		DelegateDefs: map[string]*core.DelegateDef{},
		PropertyDefs: map[string]*core.TypeDef{
			"query": {
				Type: "string",
			},
			"queryParams": {
				Type: "stream",
				Stream: &core.TypeDef{
					Type: "string",
				},
			},
			"rowColumns": {
				Type: "stream",
				Stream: &core.TypeDef{
					Type: "string",
				},
			},
			"driver": {
				Type: "string",
			},
			"url": {
				Type: "string",
			},
		},
	},
	opFunc: func(op *core.Operator) {
		query := op.Property("query").(string)

		driver := op.Property("driver").(string)
		url := op.Property("url").(string)

		params := []string{}
		for _, param := range op.Property("queryParams").([]interface{}) {
			params = append(params, param.(string))
		}

		rowColumns := []string{}
		for _, col := range op.Property("rowColumns").([]interface{}) {
			rowColumns = append(rowColumns, col.(string))
		}

		db, err := sql.Open(driver, url)
		if err != nil {
			panic(err.Error())
		}
		defer db.Close()

		err = db.Ping()
		if err != nil {
			panic(err.Error())
		}

		stmt, err := db.Prepare(query)
		if err != nil {
			panic(err.Error())
		}
		defer stmt.Close()

		in := op.Main().In()
		out := op.Main().Out()
		for !op.CheckStop() {
			i := in.Pull()
			if core.IsMarker(i) {
				out.Push(i)
				continue
			}

			im := i.(map[string]interface{})

			args := []interface{}{}
			for _, param := range params {
				args = append(args, im[param])
			}
			rows, err := stmt.Query(args...)

			if err != nil {
				out.Push(nil)
				continue
			}

			colTypes, _ := rows.ColumnTypes()
			out.PushBOS()
			for rows.Next() {
				row := make(map[string]interface{})
				row["trigger"] = nil
				dests := []interface{}{}
				for i := range rowColumns {
					colType := colTypes[i]
					var colPtr interface{}
					typeName := colType.DatabaseTypeName()
					switch typeName {
					case "VARCHAR", "TEXT", "LONGTEXT":
						colPtr = new(string)
					case "TINYINT", "SMALLINT", "MEDIUMINT", "INT", "BIGINT", "DECIMAL", "FLOAT", "DOUBLE":
						colPtr = new(float64)
					default:
						colPtr = new(string)
					}
					dests = append(dests, colPtr)
				}
				rows.Scan(dests...)
				for i, col := range rowColumns {
					row[col] = reflect.ValueOf(dests[i]).Elem().Interface()
				}
				out.Stream().Push(row)
			}
			out.PushEOS()
		}
	},
}
