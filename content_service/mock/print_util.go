package mock

import (
	"content_service/env"
	"content_service/schema"
	"fmt"
	"github.com/jinzhu/gorm"
)

func PrintSlice(s []string) {
	fmt.Printf("[")
	for _, v := range s {
		fmt.Printf("%v\n", v)
	}
	fmt.Println("]")
}

func PrintMap(m map[string]string) {
	for k, v := range m {
		fmt.Printf("%v:%v\n", k, v)
	}
	fmt.Println()
}

func PrintSet(m map[string]bool) {
	for k, v := range m {
		if v == true {
			fmt.Printf("%v\n", k)
		}
	}
	fmt.Println()
}

func PrintServiceModelMap(m map[string]map[string]schema.ModelHistory) {
	for k, v := range m {
		fmt.Printf("%v:\n", k)
		for _, v2 := range v {
			fmt.Printf("\t%v\n", v2)
		}
	}
	fmt.Println()
}

func PrintModelsByServiceMap(m []*schema.Service) {
	for _, v := range m {
		fmt.Printf("\t%v\n", v)
	}
	fmt.Println()
}

func Print_db_model_histories(db_model_histories []schema.ModelHistory) {
	fmt.Println("id\tmodel_name\ttimestamp\tmd5\tis_locked\tdesc")
	fmt.Println("----------------------------------------")
	for _, v := range db_model_histories {
		fmt.Printf("%v\n", v)
	}
	fmt.Println()
}

func PrintTable(db *gorm.DB, name string) {
	switch name {
	case "hosts":
		{
			fmt.Printf("table %v\n", name)
			fmt.Println("id\tip\tdata_center\tdesc")
			fmt.Println("----------------------------------------")
			var rows []schema.Host
			db.Find(&rows)
			for _, row := range rows {
				fmt.Printf("%v\n", row)
			}
		}
	case "services":
		{
			fmt.Printf("table %v\n", name)
			fmt.Println("id\tname\tdesc")
			fmt.Println("----------------------------------------")
			var rows []schema.Service
			db.Find(&rows)
			for _, row := range rows {
				fmt.Printf("%v\n", row)
			}
		}
	case "models":
		{
			fmt.Printf("table %v\n", name)
			fmt.Println("id\tname\tpath\tdesc")
			fmt.Println("----------------------------------------")
			var rows []schema.Model
			db.Find(&rows)
			for _, row := range rows {
				fmt.Printf("%v\n", row)
			}
		}
	case "host_services":
		{
			fmt.Printf("table %v\n", name)
			fmt.Println("id\thid\tsid\tload_weight\tdesc")
			fmt.Println("----------------------------------------")
			var rows []schema.HostService
			db.Find(&rows)
			for _, row := range rows {
				fmt.Printf("%v\n", row)
			}
		}
	case "service_models":
		{
			fmt.Printf("table %v\n", name)
			fmt.Println("id\tsid\tmid\tdesc")
			fmt.Println("----------------------------------------")
			var rows []schema.ServiceModel
			db.Find(&rows)
			for _, row := range rows {
				fmt.Printf("%v\n", row)
			}
		}
	case "model_histories":
		{
			fmt.Printf("table %v\n", name)
			fmt.Println("id\tmodel_name\ttimestamp\tmd5\tis_locked\tdesc")
			fmt.Println("----------------------------------------")
			var rows []schema.ModelHistory
			db.Find(&rows)
			for _, row := range rows {
				fmt.Printf("%v\n", row)
			}
		}
	}
	fmt.Println()
}

func PrintAllTables(env *env.Env) {
	db := env.Db
	PrintTable(db, "hosts")
	PrintTable(db, "services")
	PrintTable(db, "models")
	PrintTable(db, "host_services")
	PrintTable(db, "service_models")
	PrintTable(db, "model_histories")
}
