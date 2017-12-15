package baremetal

import (
	"fmt"
	"bytes"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	clusterv1 "k8s.io/kube-deploy/cluster-api/api/cluster/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kops/machine-controller/pkg/util"
)

// Long term, we should retrieve the current status by asking k8s, gce etc. for all the needed info.
// For now, it is stored in the matching CRD under an annotation. This is similar to
// the spec and status concept where the machine CRD is the instance spec and the annotation is the instance status.

const InstanceStatusAnnotationKey = "instance-status"

type instanceStatus *clusterv1.Machine

// Get the status of the instance identified by the given machine
func (a *BaremetalActuator) instanceStatus(machine *clusterv1.Machine) (instanceStatus, error) {
	currentMachine, err := util.GetCurrentMachineIfExists(a.machineClient, machine)
	if err != nil {
		return nil, err
	}


	if currentMachine == nil {
		// The current status no longer exists because the matching CRD has been deleted (or does not exist yet ie. bootstrapping)
		return nil, nil
	}
	return a.machineInstanceStatus(currentMachine)
}

// Sets the status of the instance identified by the given machine to the given machine
func (a *BaremetalActuator) updateInstanceStatus(machine *clusterv1.Machine) error {
	status := instanceStatus(machine)
	currentMachine, err := util.GetCurrentMachineIfExists(a.machineClient, machine)
	if err != nil {
		return err
	}

	if currentMachine == nil {
		// The current status no longer exists because the matching CRD has been deleted.
		return fmt.Errorf("Machine has already been deleted. Cannot update current instance status for machine %v", machine.ObjectMeta.Name)
	}

	m, err := a.setMachineInstanceStatus(currentMachine, status)
	if err != nil {
		return err
	}

	_, err = a.machineClient.Update(m)
	return err
}

// Gets the state of the instance stored on the given machine CRD
func (a *BaremetalActuator) machineInstanceStatus(machine *clusterv1.Machine) (instanceStatus, error) {
	if machine.ObjectMeta.Annotations == nil {
		// No state
		return nil, nil
	}

	annotation := machine.ObjectMeta.Annotations[InstanceStatusAnnotationKey]
	if annotation == "" {
		// No state
		return nil, nil
	}

	serializer := json.NewSerializer(json.DefaultMetaFactory, a.scheme, a.scheme, false)
	var status clusterv1.Machine
	_, _, err := serializer.Decode([]byte(annotation), &schema.GroupVersionKind{Group:"", Version:"cluster.k8s.io/v1alpha1", Kind:"Machine"}, &status)
	if err != nil {
		return nil, fmt.Errorf("decoding failure: %v", err)
	}

	return instanceStatus(&status), nil
}

// Applies the state of an instance onto a given machine CRD
func (a *BaremetalActuator) setMachineInstanceStatus(machine *clusterv1.Machine, status instanceStatus)  (*clusterv1.Machine, error)  {
	// Avoid status within status within status ...
	status.ObjectMeta.Annotations[InstanceStatusAnnotationKey] = ""

	serializer := json.NewSerializer(json.DefaultMetaFactory, a.scheme, a.scheme, false)
	b := []byte{}
	buff := bytes.NewBuffer(b)
	err := serializer.Encode((*clusterv1.Machine)(status), buff)
	if err != nil {
		return nil, fmt.Errorf("encoding failure: %v", err)
	}

	if machine.ObjectMeta.Annotations == nil {
		machine.ObjectMeta.Annotations = make(map[string]string)
	}
	machine.ObjectMeta.Annotations[InstanceStatusAnnotationKey] = buff.String()
	return machine, nil
}
