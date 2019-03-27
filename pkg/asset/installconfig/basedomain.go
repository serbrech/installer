package installconfig

import (
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	survey "gopkg.in/AlecAivazis/survey.v1"

	"github.com/openshift/installer/pkg/asset"
	"github.com/openshift/installer/pkg/asset/installconfig/aws"
	"github.com/openshift/installer/pkg/validate"
)

type baseDomain struct {
	BaseDomain string
}

var _ asset.Asset = (*baseDomain)(nil)

// Dependencies returns no dependencies.
func (a *baseDomain) Dependencies() []asset.Asset {
	return []asset.Asset{
		&platform{},
	}
}

// Generate queries for the base domain from the user.
func (a *baseDomain) Generate(parents asset.Parents) error {
	platform := &platform{}
	parents.Get(platform)
	platformName := platform.Name()
	switch platformName {
		case aws.Name {
			var err error
			a.BaseDomain, err = aws.GetBaseDomain()
			cause := errors.Cause(err)
			if !(aws.IsForbidden(cause) || request.IsErrorThrottle(cause)) {
				return err
			}
			logrus.Error(err)
		}
		case azure.Name:
			platform.
		case libvirt.Name, none.Name, openstack.Name, 
		default:{
			err = fmt.Errorf("unknown platform type %q", platform)
		}
	}

	return survey.Ask([]*survey.Question{
		{
			Prompt: &survey.Input{
				Message: "Base Domain",
				Help:    "The base domain of the cluster. All DNS records will be sub-domains of this base and will also include the cluster name.",
			},
			Validate: survey.ComposeValidators(survey.Required, func(ans interface{}) error {
				return validate.DomainName(ans.(string), true)
			}),
		},
	}, &a.BaseDomain)
}

// Name returns the human-friendly name of the asset.
func (a *baseDomain) Name() string {
	return "Base Domain"
}
