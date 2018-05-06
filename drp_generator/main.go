package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/digitalrebar/provision/models"
	"github.com/hashicorp/terraform/helper/schema"
)

type schemaGen struct {
	DataSource         interface{}
	Resource           interface{}
	Prefix             string
	KeyName            string
	Filename           string
	VariableName       string
	ResourceName       string
	UpdateResourceData string
	BuildModel         string
	NewImports         []string
}

func mergeLists(a1, a2 []string) []string {
	answer := map[string]bool{}
	for _, a := range a1 {
		answer[a] = true
	}
	for _, a := range a2 {
		answer[a] = true
	}

	astring := []string{}
	for k, _ := range answer {
		astring = append(astring, k)
	}
	return astring
}

func main() {
	pkgName := "drp"
	schemas := []schemaGen{}

	for _, m := range models.All() {
		pref := m.Prefix()

		// These are generally read-only.  preferences is the one to come.
		if pref == "preferences" || pref == "plugin_providers" ||
			pref == "interfaces" || pref == "jobs" || pref == "leases" {
			continue
		}

		spref := strings.TrimRight(pref, "s")
		if pref == "machines" {
			// Machine is already added, add raw_machine to manipulate raw machine objects
			spref = "raw_machine"
		}

		r := buildSchemaFromObject(m, true)
		ds := buildSchemaFromObject(m, false)

		uu, ni1 := updateResourceData(m)
		mm, ni2 := buildModel(m)
		ni := mergeLists(ni1, ni2)

		schemas = append(schemas, schemaGen{
			Prefix:             pref,
			KeyName:            m.KeyName(),
			ResourceName:       fmt.Sprintf("drp_%s", spref),
			Resource:           r,
			DataSource:         ds,
			VariableName:       fmt.Sprintf("%s", pref),
			UpdateResourceData: uu,
			BuildModel:         mm,
			NewImports:         ni,
		})
	}

	for _, s := range schemas {
		log.Printf("Generating %q...\n", s.ResourceName)
		f, err := os.Create(fmt.Sprintf("resource_%s.go", s.ResourceName))
		defer f.Close()
		if err != nil {
			log.Fatal(err)
		}

		f2, err := os.Create(fmt.Sprintf("data_source_%s.go", s.ResourceName))
		defer f.Close()
		if err != nil {
			log.Fatal(err)
		}

		resourceFields := map[string]string{}
		for k, sch := range s.Resource.(map[string]*schema.Schema) {
			resourceFields[k], _ = schemaCode(sch, "", false)
		}
		dataSourceFields := map[string]string{}
		for k, sch := range s.DataSource.(map[string]*schema.Schema) {
			dataSourceFields[k], _ = schemaCode(sch, "", false)
		}

		err = resTemplate.Execute(f, struct {
			PkgName            string
			VariableName       string
			Prefix             string
			KeyName            string
			ResourceName       string
			DataSourceFields   map[string]string
			ResourceFields     map[string]string
			UpdateResourceData string
			BuildModel         string
			NewImports         []string
		}{
			PkgName:            pkgName,
			VariableName:       s.VariableName,
			Prefix:             s.Prefix,
			KeyName:            s.KeyName,
			ResourceName:       s.ResourceName,
			DataSourceFields:   dataSourceFields,
			ResourceFields:     resourceFields,
			UpdateResourceData: s.UpdateResourceData,
			BuildModel:         s.BuildModel,
			NewImports:         s.NewImports,
		})
		if err != nil {
			log.Fatal(err)
		}

		err = dsTemplate.Execute(f2, struct {
			PkgName            string
			VariableName       string
			Prefix             string
			KeyName            string
			ResourceName       string
			DataSourceFields   map[string]string
			ResourceFields     map[string]string
			UpdateResourceData string
			BuildModel         string
			NewImports         []string
		}{
			PkgName:            pkgName,
			VariableName:       s.VariableName,
			Prefix:             s.Prefix,
			KeyName:            s.KeyName,
			ResourceName:       s.ResourceName,
			DataSourceFields:   dataSourceFields,
			ResourceFields:     resourceFields,
			UpdateResourceData: s.UpdateResourceData,
			BuildModel:         s.BuildModel,
			NewImports:         s.NewImports,
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}

var resTemplate = template.Must(template.New("res").Parse(`package {{.PkgName}}

import (
	"fmt"
	"log"

	"github.com/digitalrebar/provision/models"
	"github.com/hashicorp/terraform/helper/schema"

{{range $imp := .NewImports}}
	"{{ $imp }}"
{{end}}
)

var {{.VariableName}}ResourceSchema = map[string]*schema.Schema{
{{range $name, $schema := .ResourceFields}}
	"{{ $name }}": {{ $schema }},{{end}}
}

var {{.VariableName}}Resource = &schema.Resource{
	Schema: {{.VariableName}}ResourceSchema,
	Create: resource{{.Prefix}}Create,
        Read:   resource{{.Prefix}}Read,
        Update: resource{{.Prefix}}Update,
        Delete: resource{{.Prefix}}Delete,
        Exists: resource{{.Prefix}}Exists,
        Importer: &schema.ResourceImporter{
                State: schema.ImportStatePassthrough,
        },
}

func init() {
	theResourcesMap["{{.ResourceName}}"] = {{.VariableName}}Resource
}

{{.BuildModel}}

{{.UpdateResourceData}}

func resource{{.Prefix}}Delete(d *schema.ResourceData, meta interface{}) error {
                cc := meta.(*Config)
                log.Printf("[DEBUG] [resource{{.Prefix}}Delete] deleting %s\n", d.Id())
                _, err := cc.session.DeleteModel("{{.Prefix}}", d.Id())
                return err
        }

func resource{{.Prefix}}Exists(d *schema.ResourceData, meta interface{}) (bool, error) {
                cc := meta.(*Config)
                log.Printf("[DEBUG] [resource{{.Prefix}}Exists] testing %s\n", d.Id())
                return cc.session.ExistsModel("{{.Prefix}}",d.Id())
        }

func resource{{.Prefix}}Update(d *schema.ResourceData, meta interface{}) error {
                cc := meta.(*Config)
                log.Printf("[DEBUG] [resource{{.Prefix}}Update] updating %s\n", d.Id())

                base, err := cc.session.GetModel("{{.Prefix}}", d.Id())
                if err != nil {
                        return err
                }

		mods, err := build{{.Prefix}}Model(base, d)
                if err != nil {
                        return err
                }

                err = cc.session.Req().PatchTo(base, mods).Params("force", "true").Do(&mods)
                if err != nil {
                        return err
                }
		return update{{.Prefix}}ResourceData(mods, d)
        }

func resource{{.Prefix}}Read(d *schema.ResourceData, meta interface{}) error {
                cc := meta.(*Config)
                log.Printf("[DEBUG] [resource{{.Prefix}}Read] reading %s\n", d.Id())

                answer, err := cc.session.GetModel("{{.Prefix}}", d.Id())
                if err != nil {
                        return err
                }

		return update{{.Prefix}}ResourceData(answer, d)
        }

func resource{{.Prefix}}Create(d *schema.ResourceData, meta interface{}) error {
                cc := meta.(*Config)
                log.Printf("[DEBUG] [resource{{.Prefix}}Create] creating\n")

		mod, _ := models.New("{{.Prefix}}")
		new, err := build{{.Prefix}}Model(mod.(models.Model), d)
                if err != nil {
                        return err
                }

                answer, err := cc.session.GetModel("{{.Prefix}}", new.Key())
                if err == nil {
                        d.SetId(answer.Key())
                        ro, ok := answer.(models.Accessor)
                        if !ok || ro.IsReadOnly() {
				return update{{.Prefix}}ResourceData(answer, d)
                        }
                        return resource{{.Prefix}}Update(d, meta)
                }

                err = cc.session.CreateModel(new)
                if err != nil {
                        return err
                }

                d.SetId(new.Key())

                return resource{{.Prefix}}Read(d, meta)
        }
`))

var dsTemplate = template.Must(template.New("ds").Parse(`package {{.PkgName}}

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

var {{.VariableName}}DataSourceSchema = map[string]*schema.Schema{
{{range $name, $schema := .DataSourceFields}}
	"{{ $name }}": {{ $schema }},{{end}}
}

var {{.VariableName}}DataSource = &schema.Resource{
	Schema: {{.VariableName}}DataSourceSchema,
        Read:   dataSource{{.Prefix}}Read,
}

func init() {
	theDataSourcesMap["{{.ResourceName}}"] = {{.VariableName}}DataSource
}

func dataSource{{.Prefix}}Read(d *schema.ResourceData, meta interface{}) error {
                cc := meta.(*Config)

                id := d.Get("{{.KeyName}}").(string)
                d.SetId(id)

                log.Printf("[DEBUG] [dataSource{{.Prefix}}Read] reading %s\n", id)

                answer, err := cc.session.GetModel("{{.Prefix}}", id)
                if err != nil {
                        return err
                }

		return update{{.Prefix}}ResourceData(answer, d)
        }
`))
