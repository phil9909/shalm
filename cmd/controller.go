package cmd

import (
	"os"

	"github.com/kramerul/shalm/controllers"
	"github.com/kramerul/shalm/pkg/shalm"
	"github.com/pkg/errors"

	shalmv1a1 "github.com/kramerul/shalm/api/v1alpha1"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "run in controller mode",
	Long:  ``,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		exit(controller())
	},
}
var (
	setupLog      = ctrl.Log.WithName("setup")
	reconcilerLog = ctrl.Log.WithName("reconciler")
)

func controller() error {

	ctrl.SetLogger(zap.Logger(true))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	err = shalmv1a1.AddToScheme(mgr.GetScheme())
	if err != nil {
		return errors.Wrap(err, "unable to add shalm scheme")
	}

	reconciler := &controllers.ShalmChartReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Log:    reconcilerLog,
		Repo:   shalm.NewRepo(),
		K8s:    func(kubeconfig string) shalm.K8s { return shalm.NewK8s() },
	}
	err = reconciler.SetupWithManager(mgr)
	if err != nil {
		return errors.Wrap(err, "unable to create controller")
	}

	err = ctrl.NewWebhookManagedBy(mgr).
		For(&shalmv1a1.ShalmChart{}).
		Complete()
	if err != nil {
		return errors.Wrap(err, "unable to create webhook")
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return errors.Wrap(err, "problem running manager")
	}
	return nil
}
