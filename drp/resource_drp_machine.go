package drp

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"gitlab.com/rackn/provision/v4/models"
)

func dataSourceMachine() *schema.Resource {
	log.Println("[DEBUG] [dataSourceMachine] Initializing data structure")

	m, _ := models.New("machine")
	r := buildSchema(m, false)
	r.Create = nil
	r.Update = nil
	r.Delete = nil
	r.Importer = nil
	r.Exists = nil

	// Machines also have filters
	r.Schema["filters"] = &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"jsonvalue": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
	return r
}

func resourceMachine() *schema.Resource {
	log.Println("[DEBUG] [resourceMachine] Initializing data structure")

	m, _ := models.New("machine")
	r := buildSchema(m, true)

	r.Create = resourceMachineCreate
	r.Update = resourceMachineUpdate
	r.Delete = resourceMachineDelete

	r.Timeouts = &schema.ResourceTimeout{
		Create: schema.DefaultTimeout(25 * time.Minute),
		Update: schema.DefaultTimeout(10 * time.Minute),
		Delete: schema.DefaultTimeout(10 * time.Minute),
	}

	// Define what the machines completion stage.
	r.Schema["completion_stage"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
	}

	// Define what the machines decommision workflow
	r.Schema["decommission_workflow"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
	}

	// Define what the machines decommision workflow
	r.Schema["decommission_icon"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
	}

	// Define what the machines decommision workflow
	r.Schema["decommission_color"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
	}

	// Define what the machines decommision stage
	r.Schema["decommission_stage"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
	}

	// Define what profiles to add and remove at destroy
	r.Schema["add_profiles"] = &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}

	// Define what the machines decommision stage
	r.Schema["pool"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
	}

	// Machines also have filters
	r.Schema["filters"] = &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"jsonvalue": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
	return r
}

func allocateMachine(cc *Config, filters []string) (*models.Machine, error) {
	for {
		machines, err := cc.session.ListModel("machine", filters...)
		if err != nil {
			return nil, err
		}
		if len(machines) == 0 {
			return nil, fmt.Errorf("No machines available")
		}

		machine := machines[0]
		merged := models.Clone(machine).(*models.Machine)
		merged.Params["terraform/allocated"] = true

		ret, err := cc.session.PatchTo(machine, merged)
		if err != nil {
			berr, ok := err.(*models.Error)
			if ok {
				// If we get a patch error, the machine was allocated while we were
				// waiting.  Try again.
				if berr.Type == "PATCH" && (berr.Code == 406 || berr.Code == 409) {
					continue
				}
			}
			return nil, err
		}
		return ret.(*models.Machine), nil
	}
}

func releaseMachine(cc *Config, uuid string, tfManaged bool) error {
	for {
		machine, err := cc.session.GetModel("machines", uuid)
		if err != nil {
			return nil
		}

		merged := models.Clone(machine).(*models.Machine)
		merged.Params["terraform/allocated"] = false
		merged.Params["terraform/managed"] = tfManaged

		_, err = cc.session.PatchTo(machine, merged)
		if err != nil {
			berr, ok := err.(*models.Error)
			if ok {
				// If we get a patch error, the machine was allocated while we were
				// waiting.  Try again.
				if berr.Type == "PATCH" && (berr.Code == 406 || berr.Code == 409) {
					continue
				}
			}
			return err
		}
		return nil
	}
}

func getMachineStatus(cc *Config, uuid string, stages []string) resource.StateRefreshFunc {
	log.Printf("[DEBUG] [getMachineStatus] Getting status of machine: %s", uuid)
	return func() (interface{}, string, error) {
		mo, err := cc.session.GetModel("machines", uuid)
		if err != nil {
			log.Printf("[ERROR] [getMachineStatus] Unable to get machine: %s\n", uuid)
			return nil, "", err
		}
		machineObject := mo.(*models.Machine)

		// 6 == done  9 == pending
		machineStatus := "6"
		if machineObject.Stage != "" {
			found := false
			for _, s := range stages {
				if s == machineObject.Stage {
					found = true
					break
				}
			}
			if !found {
				machineStatus = "9"
			}
		} else {
			if machineObject.BootEnv != "local" {
				machineStatus = "9"
			}
		}

		var statusRetVal bytes.Buffer
		statusRetVal.WriteString(machineStatus)
		statusRetVal.WriteString(":")

		return machineObject, statusRetVal.String(), nil
	}
}

func machineDo(cc *Config, uuid, action string) error {
	log.Printf("[DEBUG] [machineDo] uuid: %s, action: %s", uuid, action)

	actionParams := map[string]interface{}{}
	var resp interface{}
	err := cc.session.Req().Post(actionParams).UrlFor("machines", uuid, "actions", action).Do(resp)
	if err != nil {
		log.Printf("[DEBUG] [machineDo] call %s:%s error = %v\n", uuid, action, err)
		return err
	}
	return nil
}

func updateMachine(cc *Config, machineObj *models.Machine, d *schema.ResourceData) (*models.Machine, error) {
	obj, e := buildModel(machineObj, d)
	if e != nil {
		log.Printf("[ERROR] [updateMachine] Unable to build model: %v\n", e)
		return nil, e
	}
	m := obj.(*models.Machine)

	// Make sure the add profiles are on the machine
	if ol, ok := d.GetOk("add_profiles"); ok {
		l := ol.([]interface{})
		for _, s := range l {
			prof := s.(string)

			found := false
			for _, t := range m.Profiles {
				if t == prof {
					found = true
					break
				}
			}

			if !found {
				m.Profiles = append(m.Profiles, prof)
			}
		}
	}

	cBootEnv := machineObj.BootEnv

	err := cc.session.Req().PatchTo(machineObj, m).Params("force", "true").Do(&m)
	if err != nil {
		log.Printf("[ERROR] [updateMachine] Unable to initialize machine: %v\n", err)
		return nil, err
	}
	machineObj = m

	if err := machineDo(cc, machineObj.UUID(), "nextbootpxe"); err != nil {
		log.Printf("[WARN] [updateMachine] Unable to mark the machine for pxe next boot: %s\n", machineObj.UUID())
	}

	// Power on and then cycle, if needed
	if err := machineDo(cc, machineObj.UUID(), "poweron"); err != nil {
		log.Printf("[WARN] [updateMachine] Unable to power on machine: %s\n", machineObj.UUID())
	}

	obj, err = cc.session.GetModel(machineObj.Prefix(), machineObj.Key())
	if err != nil {
		log.Printf("[ERROR] [updateMachine] Unable to re-get machine: %v\n", err)
		return nil, err
	}
	machineObj = obj.(*models.Machine)

	if machineObj.BootEnv != cBootEnv {
		if err := machineDo(cc, machineObj.Key(), "powercycle"); err != nil {
			log.Printf("[WARN] [updateMachine] Unable to power cycleup machine: %s\n", machineObj.UUID())
		}
	}

	return machineObj, nil
}

// This function doesn't really *create* a new machine but,
// consume an already registered machine.
func resourceMachineCreate(d *schema.ResourceData, meta interface{}) error {
	log.Println("[DEBUG] [resourceMachineCreate] Launching new drp_machine")
	cc := meta.(*Config)

	filters := []string{}
	if pval, set := d.GetOk("filters"); set {
		for _, o := range pval.([]interface{}) {
			v := o.(map[string]interface{})
			filters = append(filters, v["name"].(string), v["jsonvalue"].(string))
		}
	}
	if fval, set := d.GetOk("pool"); set {
		filters = append(filters, "terraform/pool", fval.(string))
	} else {
		filters = append(filters, "terraform/pool", "default")
	}
	filters = append(filters, "terraform/allocated", "false")
	filters = append(filters, "terraform/managed", "true")

	machineObj, err := allocateMachine(cc, filters)
	if err != nil {
		log.Printf("[ERROR] [resourceMachineCreate] Unable to allocate machine: %v\n", err)
		return err
	}

	uuid := machineObj.UUID()
	machineObj, err = updateMachine(cc, machineObj, d)
	if err != nil {
		log.Printf("[ERROR] [resourceMachineCreate] Unable to update machine: %v\n", err)
		if err2 := releaseMachine(cc, uuid, true); err2 != nil {
			log.Printf("[ERROR] [resourceMachineCreate] Unable to release machine: %v\n", err2)
		}
		return err
	}

	log.Printf("[DEBUG] [resourceMachineCreate] Waiting for machine (%s) to become active\n", machineObj.UUID())

	stages := []string{"complete", "complete-nowait"}
	if ns, ok := d.GetOk("completion_stage"); ok {
		stages = []string{ns.(string)}
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"9:"},
		Target:     []string{"6:"},
		Refresh:    getMachineStatus(cc, machineObj.UUID(), stages),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		if err2 := releaseMachine(cc, machineObj.UUID(), true); err2 != nil {
			log.Printf("[ERROR] [resourceMachineCreate] Unable to release machine: %v\n", err2)
		}
		return fmt.Errorf(
			"[ERROR] [resourceMachineCreate] Error waiting for machine (%s) to become deployed: %s",
			machineObj.UUID(), err)
	}

	d.SetId(machineObj.UUID())

	answer, err := cc.session.GetModel(machineObj.Prefix(), d.Id())
	if err != nil {
		return err
	}
	return updateResourceData(answer, d)
}

func resourceMachineUpdate(d *schema.ResourceData, meta interface{}) error {
	cc := meta.(*Config)
	log.Printf("[DEBUG] [resourceMachineUpdate] Modifying machine %s\n", d.Id())

	obj, err := cc.session.GetModel("machines", d.Id())
	if err != nil {
		log.Printf("[ERROR] [resourceMachineUpdate] Unable to get machine: %v\n", err)
		return err
	}
	machineObj := obj.(*models.Machine)

	machineObj, err = updateMachine(cc, machineObj, d)
	if err != nil {
		log.Printf("[ERROR] [resourceMachineCreate] Unable to update machine: %v\n", err)
		return err
	}

	log.Printf("[DEBUG] Done Modifying machine %s\n", d.Id())
	return updateResourceData(machineObj, d)
}

// This function doesn't really *delete* a drp managed machine but releases (read, turns off) the machine.
func resourceMachineDelete(d *schema.ResourceData, meta interface{}) error {
	cc := meta.(*Config)
	log.Printf("[DEBUG] Deleting machine %s\n", d.Id())

	obj, err := cc.session.GetModel("machines", d.Id())
	if err != nil {
		log.Printf("[ERROR] [resourceMachineDelete] Failed to get machine: %v\n", err)
		return err
	}
	machineObj := obj.(*models.Machine)
	newObj := models.Clone(machineObj).(*models.Machine)

	if nc, ok := d.GetOk("decommission_color"); ok {
		newObj.Meta["color"] = nc.(string)
	} else {
		newObj.Meta["color"] = "black"
	}

	if ni, ok := d.GetOk("decommission_icon"); ok {
		newObj.Meta["icon"] = ni.(string)
	} else {
		newObj.Meta["icon"] = "map outline"
	}

	if nw, ok := d.GetOk("decommission_workflow"); ok {
		newObj.Workflow = nw.(string)
	} else {
		if ns, ok := d.GetOk("decommission_stage"); ok {
			newObj.Stage = ns.(string)
		} else {
			if machineObj.Workflow != "" {
				newObj.Workflow = "discover"
			} else {
				if machineObj.Stage != "" {
					newObj.Stage = "discover"
				} else {
					newObj.BootEnv = "sledgehammer"
				}
			}
		}
		// Runnable should only be set to false if we aren't using workflows.
		newObj.Runnable = false
	}

	// Remove the profiles
	if ol, ok := d.GetOk("add_profiles"); ok {
		l := ol.([]interface{})
		newList := []string{}

		for _, ts := range newObj.Profiles {
			found := false
			for _, s := range l {
				prof := s.(string)
				if prof == ts {
					found = true
					break
				}
			}
			if !found {
				newList = append(newList, ts)
			}
		}

		newObj.Profiles = newList
	}

	// Update the machine to request position
	err = cc.session.Req().PatchTo(machineObj, newObj).Params("force", "true").Do(&newObj)
	if err != nil {
		log.Printf("[ERROR] [resourceMachineDelete] Unable to reset machine: %v\n", err)
		return err
	}

	if err := releaseMachine(cc, d.Id(), false); err != nil {
		return err
	}

	if err := machineDo(cc, machineObj.UUID(), "nextbootpxe"); err != nil {
		log.Printf("[ERROR] [resourceMachineRelease] Unable to mark the machine for pxe next boot: %s\n", machineObj.UUID())
	}

	if err := machineDo(cc, machineObj.UUID(), "powercycle"); err != nil {
		log.Printf("[ERROR] [resourceMachineRelease] Unable to power cycle machine: %s\n", machineObj.UUID())
	}

	log.Printf("[DEBUG] [resourceMachineDelete] Machine (%s) released", d.Id())

	d.SetId("")

	return nil
}
