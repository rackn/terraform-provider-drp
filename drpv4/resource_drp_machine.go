package drpv4

/*
 * Copyright RackN 2020
 */

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gitlab.com/rackn/provision/v4/models"
)

func resourceMachine() *schema.Resource {
	r := &schema.Resource{
		Create: resourceMachineAllocate,
		Read:   resourceMachineRead,
		Update: resourceMachineUpdate,
		Delete: resourceMachineRelease,

		Schema: map[string]*schema.Schema{

			"pool": &schema.Schema{
				Type:        schema.TypeString,
				Default:     "default",
				Description: "Pool to operate for machine actions (Machine.Pool)",
				ForceNew:    true,
				Optional:    true,
			},
			"timeout": &schema.Schema{
				Type:        schema.TypeString,
				Default:     "5m",
				Description: "Pooling Request: max time string to wait for pool operations",
				Optional:    true,
			},
			"add_profiles": &schema.Schema{
				Type:        schema.TypeList,
				Description: "Pooling Request: Profiles to add to Machine.Profiles (must already exist)",
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			// sets parameters to add
			"add_parameters": &schema.Schema{
				Type:        schema.TypeList,
				Description: "Pooling Request: Parameters (key: value) to add to Machine.Params",
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			// sets parameters to add
			"filters": &schema.Schema{
				Type:        schema.TypeList,
				Description: "Pooling Request: Selection Filters (uses Digital Rebar format e.g. FilterVar=value)",
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"authorized_keys": &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Pooling Request: Sets access-keys param on machine requested",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			// Machine.Address
			"address": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Returns Digital Rebar Machine.Address",
				Computed:    true,
			},
			// Machine.Status
			"status": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Returns Digital Rebar Machine.Status",
				Computed:    true,
			},
			// Machine.Name
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "Returns Digital Rebar Machine.Name",
				Computed:    true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(25 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}

	return r
}

func resourceMachineAllocate(d *schema.ResourceData, m interface{}) error {
	log.Println("[DEBUG] [resourceMachineAllocate] Allocating new drp_machine")
	cc := m.(*Config)

	pool := d.Get("pool").(string)
	if pool == "" {
		pool = "default"
	}
	d.Set("pool", pool)
	timeout := d.Get("timeout").(string)
	parms := map[string]interface{}{
		"pool/wait-timeout": timeout,
	}
	d.Set("timeout", timeout)

	if profiles, ok := d.GetOk("add_profiles"); ok {
		parms["pool/add-profiles"] = profiles.([]interface{})
	}

	parameters := map[string]interface{}{}
	if ap, ok := d.GetOk("authorized_keys"); ok {
		akeys := ap.([]interface{})
		accesskeys := map[string]string{}
		for i, p := range akeys {
			accesskeys[fmt.Sprintf("terraform-%d", i)] = p.(string)
		}
		parameters["access-keys"] = accesskeys
	}
	if ap, ok := d.GetOk("add_parameters"); ok {
		aparams := ap.([]interface{})
		for _, p := range aparams {
			param := strings.Split(p.(string), ":")
			if len(param) < 2 {
				return fmt.Errorf("Error in add_parameter format: %+v", aparams)
			}
			key := param[0]
			value := strings.TrimLeft(param[1], " ")
			parameters[key] = value
		}
	}
	if len(parameters) > 0 {
		parms["pool/add-parameters"] = parameters
	}

	if filters, ok := d.GetOk("filters"); ok {
		parms["pool/filter"] = filters.([]interface{})
	}

	pr := []*models.PoolResult{}
	req := cc.session.Req().Post(parms).UrlFor("pools", pool, "allocateMachines")
	if err := req.Do(&pr); err != nil {
		log.Printf("[DEBUG] POST error %+v | %+v", err, req)
		return fmt.Errorf("Error allocated from pool %s: %s", pool, err)
	}
	mc := pr[0]
	log.Printf("[DEBUG] Allocated %s machine %s (%s)", mc.Status, mc.Name, mc.Uuid)
	d.Set("status", mc.Status)
	d.Set("name", mc.Name)
	d.SetId(mc.Uuid)
	return resourceMachineRead(d, m)
}

func resourceMachineRead(d *schema.ResourceData, m interface{}) error {
	log.Println("[DEBUG] [resourceMachineRead] Reading drp_machine")
	cc := m.(*Config)

	uuid := d.Id()
	if uuid == "" {
		return fmt.Errorf("Requires Uuid from id")
	}

	log.Printf("[DEBUG] Reading machine %s", uuid)
	mo, err := cc.session.GetModel("machines", uuid)
	if err != nil {
		log.Printf("[ERROR] [resourceMachineRead] Unable to get machine: %s", uuid)
		return fmt.Errorf("Unable to get machine %s", uuid)
	}
	machineObject := mo.(*models.Machine)

	d.Set("status", machineObject.PoolStatus)
	d.Set("address", machineObject.Address)
	d.Set("name", machineObject.Name)

	return nil
}

func resourceMachineUpdate(d *schema.ResourceData, m interface{}) error {
	log.Println("[DEBUG] [resourceMachineUpdate] Updating drp_machine")
	cc := m.(*Config)

	// at this time there are no updates
	log.Printf("[DEBUG] Config %v", cc)

	return resourceMachineRead(d, m)
}

func resourceMachineRelease(d *schema.ResourceData, m interface{}) error {
	log.Println("[DEBUG] [resourceMachineRelease] Releasing drp_machine")
	cc := m.(*Config)

	uuid := d.Id()
	if uuid == "" {
		return fmt.Errorf("Requires Uuid from id")
	}
	pool := d.Get("pool").(string)
	if pool == "" {
		return fmt.Errorf("Requires Pool")
	}
	log.Printf("[DEBUG] Releasing %s from %s", uuid, pool)

	pr := []*models.PoolResult{}
	parms := map[string]interface{}{
		"pool/wait-timeout": d.Get("timeout").(string),
		"pool/machine-list": []string{uuid},
	}
	// remove the added profiles
	if profiles, ok := d.GetOk("add_profiles"); ok {
		parms["pool/remove-profiles"] = profiles.([]interface{})
	}

	parameters := []string{}
	if _, ok := d.GetOk("authorized_keys"); ok {
		parameters = append(parameters, "access-keys")
	}
	if ap, ok := d.GetOk("add_parameters"); ok {
		aparams := ap.([]interface{})
		for _, p := range aparams {
			param := strings.Split(p.(string), ":")
			if len(param) < 2 {
				return fmt.Errorf("Error in add_parameter format: %+v", aparams)
			}
			key := param[0]
			parameters = append(parameters, key)
		}
	}
	if len(parameters) > 0 {
		parms["pool/remove-parameters"] = parameters
	}

	req := cc.session.Req().Post(parms).UrlFor("pools", pool, "releaseMachines")
	if err := req.Do(&pr); err != nil {
		log.Printf("[DEBUG] POST error %+v | %+v", err, req)
		return fmt.Errorf("Error releasing %s from pool %s: %s", uuid, pool, err)
	}

	mc := pr[0]
	if mc.Status == "Free" {
		d.Set("status", mc.Status)
		d.Set("address", "")
		d.Set("name", uuid)
		d.SetId("")
		return nil
	} else {
		return fmt.Errorf("Could not release %s from pool %s", uuid, pool)
	}

}
