package drp

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/VictorLowther/jsonpatch2/utils"
	"github.com/digitalrebar/provision/models"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pborman/uuid"
)

func buildSchemaListFromObject(m interface{}) *schema.Schema {
	r := &schema.Resource{
		Schema: buildSchemaFromObject(m),
	}
	return &schema.Schema{
		Type:     schema.TypeList,
		Elem:     r,
		Optional: true,
		Computed: true,
	}
}

func buildSchemaFromObject(m interface{}) map[string]*schema.Schema {
	sm := map[string]*schema.Schema{}

	val := reflect.ValueOf(m).Elem()

	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		// Skip non-exported fields
		if typeField.PkgPath != "" {
			continue
		}

		// Skip the access and validation fields
		if typeField.Name == "Access" || typeField.Name == "Validation" {
			continue
		}

		// Skip the Profile - deprecated fields
		if typeField.Name == "Profile" {
			continue
		}

		fieldName := typeField.Name
		// Provider is reserved Terraform name
		if fieldName == "Provider" {
			fieldName = "PluginProvider"
		}

		// Meta is a constant map of strings (but shows up as a type of Meta - fix it)
		if fieldName == "Meta" {
			sm["Meta"] = &schema.Schema{
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				Computed: true,
			}

			continue
		}

		//
		// This is a cluster.  Terraform doesn't generic interface{}
		// basically, interface{} and map[string]interface{}
		//
		// Will try some things.
		//
		if fieldName == "Params" {
			// GREG: FIGURE THIS OUT!!!
			continue
		}
		if fieldName == "Schema" {
			// GREG: FIGURE THIS OUT!!!
			continue
		}

		if strings.HasPrefix(typeField.Type.String(), "[]") {
			listType := typeField.Type.String()[2:]

			switch listType {
			case "string":
				sm[fieldName] = &schema.Schema{
					Type: schema.TypeList,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Optional: true,
					Computed: true,
				}
			case "models.DhcpOption", "*models.DhcpOption":
				sm[fieldName] = buildSchemaListFromObject(&models.DhcpOption{})

			case "models.TemplateInfo":
				sm[fieldName] = buildSchemaListFromObject(&models.TemplateInfo{})
			case "uint8":
				sm[fieldName] = &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				}
			default:
				fmt.Printf("[DEBUG] UNKNOWN List Field Name: %s (%s),\t Tag Value: %s\n",
					fieldName, typeField.Type,
					tag.Get("tag_name"))
			}
			continue
		}

		switch typeField.Type.String() {
		case "models.OsInfo":
			// Singleton struct - encode as a list for now.
			sm[fieldName] = buildSchemaListFromObject(&models.OsInfo{})
		case "string", "net.IP", "uuid.UUID", "time.Time":
			sm[fieldName] = &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			}
		case "bool":
			sm[fieldName] = &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			}
		case "int", "int32", "uint8":
			sm[fieldName] = &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			}
		default:
			fmt.Printf("[DEBUG] UNKNOWN Base Field Name: %s (%s),\t Tag Value: %s\n",
				fieldName, typeField.Type,
				tag.Get("tag_name"))
		}
	}

	return sm
}

func resourceGeneric(pref string) *schema.Resource {
	log.Printf("[DEBUG] [resourceGeneric] Initializing data structure: %s\n", pref)
	m, _ := models.New(pref)
	return buildSchema(m)
}

func buildSchema(m models.Model) *schema.Resource {
	r := &schema.Resource{
		Create: createDefaultCreateFunction(m),
		Read:   createDefaultReadFunction(m),
		Update: createDefaultUpdateFunction(m),
		Delete: createDefaultDeleteFunction(m),
		Exists: createDefaultExistsFunction(m),
		Schema: buildSchemaFromObject(m),
	}

	return r
}

func updateResourceData(m models.Model, d *schema.ResourceData) error {
	val := reflect.ValueOf(m).Elem()

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		// Skip the access and validation fields
		if typeField.Name == "Access" || typeField.Name == "Validation" {
			continue
		}

		// Skip the Profile - deprecated fields
		if typeField.Name == "Profile" {
			continue
		}

		fieldName := typeField.Name
		// Provider is reserved Terraform name
		if fieldName == "Provider" {
			fieldName = "PluginProvider"
		}

		// Meta is a constant map of strings (but shows up as a type of Meta - fix it)
		if fieldName == "Meta" {
			d.Set("Meta", valueField.Interface())
			continue
		}

		//
		// This is a cluster.  Terraform doesn't generic interface{}
		// basically, interface{} and map[string]interface{}
		//
		// Will try some things.
		//
		if fieldName == "Params" {
			// GREG: FIGURE THIS OUT!!!
			fmt.Printf("[DEBUG] Params not support for push into terraform\n")
			continue
		}
		if fieldName == "Schema" {
			// GREG: FIGURE THIS OUT!!!
			fmt.Printf("[DEBUG] Schema not support for push into terraform\n")
			continue
		}

		if strings.HasPrefix(typeField.Type.String(), "[]") {
			listType := typeField.Type.String()[2:]

			switch listType {
			case "string", "*models.DhcpOption", "models.DhcpOption", "models.TemplateInfo":
				d.Set(fieldName, valueField.Interface())
			case "uint8":
				d.Set(fieldName, fmt.Sprintf("%s", valueField.Interface()))
			default:
				log.Printf("[DEBUG] UNKNOWN Field Name: %s (%s),\t Field Value: %v,\t Tag Value: %s\n",
					fieldName, typeField.Type,
					valueField.Interface(), tag.Get("tag_name"))
			}
			continue
		}

		switch typeField.Type.String() {
		case "models.OsInfo":
			d.Set(fieldName, []models.OsInfo{valueField.Interface().(models.OsInfo)})
		case "string", "net.IP", "uuid.UUID", "time.Time":
			d.Set(fieldName, fmt.Sprintf("%s", valueField.Interface()))
		case "bool":
			d.Set(fieldName, valueField.Interface())
		case "int", "int32", "uint8":
			d.Set(fieldName, valueField.Interface())
		default:
			log.Printf("[DEBUG] UNKNOWN Field Name: %s (%s),\t Field Value: %v,\t Tag Value: %s\n",
				fieldName, typeField.Type,
				valueField.Interface(), tag.Get("tag_name"))
		}
	}
	return nil
}

func buildModel(m models.Model, d *schema.ResourceData) (models.Model, error) {
	new := models.Clone(m)

	val := reflect.ValueOf(new).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		// Skip the access and validation fields
		if typeField.Name == "Access" || typeField.Name == "Validation" {
			continue
		}

		// Skip the Profile - deprecated fields
		if typeField.Name == "Profile" {
			continue
		}

		fieldName := typeField.Name
		// Provider is reserved Terraform name
		if fieldName == "Provider" {
			fieldName = "PluginProvider"
		}

		if !d.HasChange(fieldName) {
			continue
		}

		// Meta is a constant map of strings (but shows up as a type of Meta - fix it)
		if fieldName == "Meta" {
			valueField.Set(reflect.MakeMap(typeField.Type))
			ms := d.Get("Meta").(map[string]interface{})
			for k, v := range ms {
				valueField.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
			}
			continue
		}

		//
		// This is a cluster.  Terraform doesn't generic interface{}
		// basically, interface{} and map[string]interface{}
		//
		// Will try some things.
		//
		if fieldName == "Params" {
			fmt.Printf("[DEBUG] Params not support for push to DRP\n")
			continue
		}
		if fieldName == "Schema" {
			// GREG: FIGURE THIS OUT!!!
			fmt.Printf("[DEBUG] Schema not support for push to DRP\n")
			continue
		}

		if strings.HasPrefix(typeField.Type.String(), "[]") {
			listType := typeField.Type.String()[2:]
			subType := typeField.Type.Elem()

			switch listType {
			case "string", "models.TemplateInfo",
				"models.DhcpOption", "*models.DhcpOption":

				data := d.Get(fieldName).([]interface{})
				v := reflect.MakeSlice(typeField.Type, 0, len(data))
				for _, s := range data {
					no := reflect.New(subType).Interface()
					if err := utils.Remarshal(s, no); err != nil {
						return nil, err
					}
					v = reflect.Append(v, reflect.Indirect(reflect.ValueOf(no)))
				}
				valueField.Set(v)

			case "uint8":
				fmt.Printf("[DEBUG] list of %s not support for push to DRP\n", listType)
			default:
				fmt.Printf("[DEBUG] UNKNOWN Field Name: %s (%s),\t Field Value: %v,\t Tag Value: %s\n",
					fieldName, typeField.Type,
					valueField.Interface(), tag.Get("tag_name"))
			}
			continue
		}

		switch typeField.Type.String() {
		case "models.OsInfo":
			data := d.Get(fieldName).([]interface{})
			for _, s := range data {
				no := models.OsInfo{}
				if err := utils.Remarshal(s, &no); err != nil {
					return nil, err
				}
				valueField.Set(reflect.ValueOf(no))
				break
			}
		case "string":
			valueField.SetString(d.Get(fieldName).(string))
		case "net.IP":
			ip := net.ParseIP(d.Get(fieldName).(string))
			valueField.Set(reflect.ValueOf(ip))
		case "uuid.UUID":
			uu := uuid.Parse(d.Get(fieldName).(string))
			valueField.Set(reflect.ValueOf(uu))
		case "time.Time":
			if t, e := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST",
				d.Get(fieldName).(string)); e != nil {
				return nil, e
			} else {
				valueField.Set(reflect.ValueOf(t))
			}
		case "bool":
			valueField.SetBool(d.Get(fieldName).(bool))
		case "int", "int32", "uint8":
			valueField.SetInt(int64(d.Get(fieldName).(int)))
		default:
			fmt.Printf("[DEBUG] UNKNOWN Field Name: %s (%s),\t Field Value: %v,\t Tag Value: %s\n",
				fieldName, typeField.Type,
				valueField.Interface(), tag.Get("tag_name"))
		}
	}
	return new, nil
}

func createDefaultCreateFunction(m models.Model) func(*schema.ResourceData, interface{}) error {
	return func(d *schema.ResourceData, meta interface{}) error {
		cc := meta.(*Config)
		log.Printf("[DEBUG] [resource%sCreate] creating\n", m.Prefix())

		new, err := buildModel(m, d)
		if err != nil {
			return err
		}

		answer, err := cc.session.GetModel(new.Prefix(), new.Key())
		if err == nil {
			d.SetId(answer.Key())
			ro, ok := answer.(models.Accessor)
			if !ok || ro.IsReadOnly() {
				return updateResourceData(answer, d)
			}
			return createDefaultUpdateFunction(m)(d, meta)
		}

		err = cc.session.CreateModel(new)
		if err != nil {
			return err
		}

		d.SetId(new.Key())

		return createDefaultReadFunction(m)(d, meta)
	}
}

func createDefaultReadFunction(m models.Model) func(*schema.ResourceData, interface{}) error {
	return func(d *schema.ResourceData, meta interface{}) error {
		cc := meta.(*Config)
		log.Printf("[DEBUG] [resource%sRead] reading %s\n", m.Prefix(), d.Id())

		answer, err := cc.session.GetModel(m.Prefix(), d.Id())
		if err != nil {
			return err
		}

		return updateResourceData(answer, d)
	}
}

func createDefaultUpdateFunction(m models.Model) func(*schema.ResourceData, interface{}) error {
	return func(d *schema.ResourceData, meta interface{}) error {
		cc := meta.(*Config)
		log.Printf("[DEBUG] [resource%sUpdate] updating %s\n", m.Prefix(), d.Id())

		base, err := cc.session.GetModel(m.Prefix(), d.Id())
		if err != nil {
			return err
		}

		mods, err := buildModel(base, d)
		if err != nil {
			return err
		}

		answer, err := cc.session.PatchTo(base, mods)
		if err != nil {
			return err
		}
		return updateResourceData(answer, d)
	}
}

func createDefaultDeleteFunction(m models.Model) func(*schema.ResourceData, interface{}) error {
	return func(d *schema.ResourceData, meta interface{}) error {
		cc := meta.(*Config)
		log.Printf("[DEBUG] [resource%sDelete] deleting %s\n", m.Prefix(), d.Id())
		_, err := cc.session.DeleteModel(m.Prefix(), d.Id())
		return err
	}
}

func createDefaultExistsFunction(m models.Model) func(*schema.ResourceData, interface{}) (bool, error) {
	return func(d *schema.ResourceData, meta interface{}) (bool, error) {
		cc := meta.(*Config)
		log.Printf("[DEBUG] [resource%sExists] testing %s\n", m.Prefix(), d.Id())
		return cc.session.ExistsModel(m.Prefix(), d.Id())
	}
}
