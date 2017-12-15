package drp

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/digitalrebar/provision/models"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceMachine() *schema.Resource {
	log.Println("[DEBUG] [resourceMachine] Initializing data structure")

	m, _ := models.New("machine")
	r := buildSchema(m)

	r.Create = resourceMachineCreate
	r.Update = resourceMachineUpdate
	r.Delete = resourceMachineDelete

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

func getMachineStatus(cc *Config, uuid string) resource.StateRefreshFunc {
	log.Printf("[DEBUG] [getMachineStatus] Getting status of machine: %s", uuid)
	return func() (interface{}, string, error) {
		mo, err := cc.session.GetModel("machines", uuid)
		if err != nil {
			log.Printf("[ERROR] [getMachineStatus] Unable to get machine: %s\n", uuid)
			return nil, "", err
		}
		machineObject := mo.(*models.Machine)

		machineStatus := "6"
		if machineObject.Stage != "" {
			if machineObject.Stage != "complete" && machineObject.Stage != "complete-nowait" {
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

	cBootEnv := machineObj.BootEnv

	err := cc.session.Req().PatchTo(machineObj, m).Params("force", "true").Do(&m)
	if err != nil {
		log.Printf("[ERROR] [updateMachine] Unable to initialize machine: %v\n", err)
		return nil, err
	}
	machineObj = m

	if err := machineDo(cc, machineObj.UUID(), "nextbootpxe"); err != nil {
		log.Printf("[ERROR] [updateMachine] Unable to mark the machine for pxe next boot: %s\n", machineObj.UUID())
		return nil, err
	}

	// Power on and then cycle, if needed
	if err := machineDo(cc, machineObj.UUID(), "poweron"); err != nil {
		log.Printf("[ERROR] [updateMachine] Unable to power on machine: %s\n", machineObj.UUID())
		return nil, err
	}

	obj, err = cc.session.GetModel(machineObj.Prefix(), machineObj.Key())
	if err != nil {
		log.Printf("[ERROR] [updateMachine] Unable to re-get machine: %v\n", err)
		return nil, err
	}
	machineObj = obj.(*models.Machine)

	if machineObj.BootEnv != cBootEnv {
		if err := machineDo(cc, machineObj.Key(), "powercycle"); err != nil {
			log.Printf("[ERROR] [updateMachine] Unable to power cycleup machine: %s\n", machineObj.UUID())
			return nil, err
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
			filters = append(filters, v["name"].(string), v["value"].(string))
		}
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
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"9:"},
		Target:     []string{"6:"},
		Refresh:    getMachineStatus(cc, machineObj.UUID()),
		Timeout:    25 * time.Minute,
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

	if machineObj.Stage != "" {
		newObj.Stage = "discover"
	} else {
		newObj.BootEnv = "sledgehammer"
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