// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package worker

import (
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/imagevector"
	"github.com/gardener/gardener-extension-provider-kubevirt/pkg/kubevirt"
	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/util"

	"github.com/gardener/gardener/extensions/pkg/controller/worker"
	"github.com/gardener/gardener/extensions/pkg/controller/worker/genericactuator"
	machinescheme "github.com/gardener/machine-controller-manager/pkg/client/clientset/versioned/scheme"
	"github.com/pkg/errors"
	apiextensionsscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	// DefaultAddOptions are the default AddOptions for AddToManager.
	DefaultAddOptions = AddOptions{}
)

// AddOptions are options to apply when adding the KubeVirt worker controller to the manager.
type AddOptions struct {
	// Controller are the controller.Options.
	Controller controller.Options
	// IgnoreOperationAnnotation specifies whether to ignore the operation annotation or not.
	IgnoreOperationAnnotation bool
}

// AddToManagerWithOptions adds a controller with the given Options to the given manager.
// The opts.Reconciler is being set with a newly instantiated actuator.
func AddToManagerWithOptions(mgr manager.Manager, opts AddOptions) error {
	scheme := mgr.GetScheme()
	if err := apiextensionsscheme.AddToScheme(scheme); err != nil {
		return err
	}
	if err := machinescheme.AddToScheme(scheme); err != nil {
		return err
	}

	dataVolumeManager, err := kubevirt.NewDefaultDataVolumeManager(kubevirt.ClientFactoryFunc(kubevirt.GetClient))
	if err != nil {
		return errors.Wrap(err, "could not create kubevirt data volume manager")
	}

	delegateFactory := &delegateFactory{
		logger:            workerLogger,
		dataVolumeManager: dataVolumeManager,
	}

	logger := log.Log.WithName("worker-actuator")

	return worker.Add(mgr, worker.AddArgs{
		Actuator: NewActuator(genericactuator.NewActuator(
			workerLogger,
			delegateFactory,
			kubevirt.MachineControllerManagerName,
			mcmChart,
			mcmShootChart,
			imagevector.ImageVector(),
			extensionscontroller.ChartRendererFactoryFunc(util.NewChartRendererForShoot)),
			delegateFactory.Client(),
			logger,
			dataVolumeManager),
		ControllerOptions: opts.Controller,
		Predicates:        worker.DefaultPredicates(opts.IgnoreOperationAnnotation),
		Type:              kubevirt.Type,
	})
}

// AddToManager adds a controller with the default Options.
func AddToManager(mgr manager.Manager) error {
	return AddToManagerWithOptions(mgr, DefaultAddOptions)
}
