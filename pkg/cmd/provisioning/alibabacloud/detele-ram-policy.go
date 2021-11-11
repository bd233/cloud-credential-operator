package alibabacloud

import (
	"fmt"
	"log"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ram"
	"github.com/openshift/cloud-credential-operator/pkg/alibabacloud"
	credreqv1 "github.com/openshift/cloud-credential-operator/pkg/apis/cloudcredential/v1"
	"github.com/openshift/cloud-credential-operator/pkg/cmd/provisioning"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	// DeleteRAMPolicyOpts captures the options that affect deletion
	// of the RAM policies.
	DeleteRAMPolicyOpts = options{}
)

// NewDeleteRAMPolicyCmd provides the "delete-ram-policy" subcommand
func NewDeleteRAMPolicyCmd() *cobra.Command {
	deleteRAMPolicyCmd := &cobra.Command{
		Use:   "delete-ram-policy",
		Short: "Delete RAM Policy",
		Run:   deleteRAMPolicyCmd,
	}

	deleteRAMPolicyCmd.PersistentFlags().StringVar(&DeleteRAMPolicyOpts.Name, "name", "", "User-define name for all created Alibaba Cloud resources (can be separate from the cluster's infra-id)")
	deleteRAMPolicyCmd.MarkPersistentFlagRequired("name")
	deleteRAMPolicyCmd.PersistentFlags().StringVar(&DeleteRAMPolicyOpts.CredRequestDir, "credentials-requests-dir", "", "Directory containing files of CredentialsRequests to create RAM AK for (can be created by running 'oc adm release extract --credentials-requests --cloud=alibabacloud' against an OpenShift release image)")
	deleteRAMPolicyCmd.MarkPersistentFlagRequired("credentials-requests-dir")
	deleteRAMPolicyCmd.PersistentFlags().StringVar(&DeleteRAMPolicyOpts.RootAccessKeyId, "root-access-key", "", "The root user ak with ram permission such as DeletePolicy")
	deleteRAMPolicyCmd.MarkPersistentFlagRequired("root-access-key")
	deleteRAMPolicyCmd.PersistentFlags().StringVar(&DeleteRAMPolicyOpts.RootAccessKeySecret, "root-access-key-secret", "", "The root user sk with ram permission such as DeletePolicy")
	deleteRAMPolicyCmd.MarkPersistentFlagRequired("root-access-key-secret")
	deleteRAMPolicyCmd.PersistentFlags().StringVar(&DeleteRAMPolicyOpts.UserName, "user-name", "", "The specific ram user name, the user has attached all permission defined in CredentialsRequests")
	deleteRAMPolicyCmd.MarkPersistentFlagRequired("user-name")
	deleteRAMPolicyCmd.PersistentFlags().StringVar(&DeleteRAMPolicyOpts.Region, "region", "", "Alibaba Cloud region endpoint only required for GovCloud")

	return deleteRAMPolicyCmd
}

func deleteRAMPolicyCmd(cmd *cobra.Command, args []string) {
	client, err := alibabacloud.NewClient(DeleteRAMPolicyOpts.Region, DeleteRAMPolicyOpts.RootAccessKeyId, DeleteRAMPolicyOpts.RootAccessKeySecret)
	if err != nil {
		log.Fatal(err)
	}
	err = deleteRAMPolicy(client, DeleteRAMPolicyOpts.Name, DeleteRAMPolicyOpts.CredRequestDir, DeleteRAMPolicyOpts.UserName)
	if err != nil {
		log.Fatal(err)
	}
}

func deleteRAMPolicy(client alibabacloud.Client, name, credReqDir string, userName string) error {
	// Process directory
	credRequests, err := provisioning.GetListOfCredentialsRequests(credReqDir)
	if err != nil {
		return errors.Wrap(err, "Failed to process files containing CredentialsRequests")
	}

	for _, cr := range credRequests {
		// infraName-targetNamespace-targetSecretName
		err := detachAndDeletePolicy(client, name, cr, userName)
		if err != nil {
			return err
		}
	}

	return nil
}

func detachAndDeletePolicy(client alibabacloud.Client, name string, credReq *credreqv1.CredentialsRequest, userName string) (err error) {
	policyName := fmt.Sprintf("%s-%s-policy", name, credReq.Spec.SecretRef.Name)

	err = detachComponentPolicy(client, userName, policyName)
	if err != nil {
		return err
	}
	log.Printf("policy %s has dttached on user %s", policyName, userName)

	err = deleteComponentPolicy(client, policyName)
	if err != nil {
		return err
	}
	log.Printf("ram policy %s has deleted", policyName)
	return nil
}

func detachComponentPolicy(client alibabacloud.Client, user string, policyName string) (err error) {
	req := ram.DetachPolicyFromUserRequest{}
	req.PolicyType = Https
	req.PolicyName = policyName
	req.PolicyType = Custom
	_, err = client.DetachPolicyFromUser(&req)
	return err
}

func deleteComponentPolicy(client alibabacloud.Client, policyName string) error {
	req := ram.DeletePolicyRequest{}
	req.PolicyName = policyName
	_, err := client.DeletePolicy(&req)
	return err
}
