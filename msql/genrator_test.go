package msql

import (
	"fmt"
	"github.com/gomelon/meta"
	"github.com/gomelon/meta-templates/msql/templates"
	"os"
	"testing"
)

func TestTemplateGen(t *testing.T) {

	workdir, _ := os.Getwd()
	path := workdir + "/testdata"
	generator, err := meta.NewTemplateGenerator(path, templates.NestSqlDB, meta.WithMetas(Metas),
		meta.WithFuncMapProvider(func(generator *meta.TemplateGenerator) map[string]any {
			return NewFunctions(generator).FuncMap()
		}))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	//err = generator.Print()
	err = generator.Generate()
	//err = generator.Generate()
	if err != nil {
		fmt.Println(err.Error())
	}
}
