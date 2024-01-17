package iampolicy

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/client/iam_service"
	iamModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-iam/stable/2019-12-10/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
)

type ResourceUpdater interface {
	// GetIamPolicy gets the existing IAM policy attached to a resource.
	GetIamPolicy(context.Context) (*models.HashicorpCloudResourcemanagerPolicy, error)

	// SetIamPolicy replaces the existing IAM Policy attached to a resource.
	SetIamPolicy(ctx context.Context, policy *models.HashicorpCloudResourcemanagerPolicy) (*models.HashicorpCloudResourcemanagerPolicy, error)
}

type Setter interface {
	SetPolicy(ctx context.Context, policy *models.HashicorpCloudResourcemanagerPolicy) (*models.HashicorpCloudResourcemanagerPolicy, error)
	AddBinding(ctx context.Context, principalID, roleID string) (*models.HashicorpCloudResourcemanagerPolicy, error)
	DeleteBinding(ctx context.Context, principalID, roleID string) (*models.HashicorpCloudResourcemanagerPolicy, error)
}

type setter struct {
	orgID   string
	updater ResourceUpdater
	iam     iam_service.ClientService
	logger  hclog.Logger
}

func NewSetter(organizationID string, updater ResourceUpdater, iam iam_service.ClientService, logger hclog.Logger) Setter {
	return &setter{
		orgID:   organizationID,
		updater: updater,
		iam:     iam,
		logger:  logger.Named("iampolicy_setter"),
	}
}

func (s *setter) SetPolicy(ctx context.Context, policy *models.HashicorpCloudResourcemanagerPolicy) (*models.HashicorpCloudResourcemanagerPolicy, error) {
	if policy == nil {
		return nil, fmt.Errorf("nil policy passed")
	}

	// If the policy is missing an etag, we need to fetch the current policy.
	if policy.Etag == "" {
		s.logger.Debug("fetching existing policy in order to populate etag")
		existing, err := s.updater.GetIamPolicy(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve existing policy: %w", err)
		}

		s.logger.Debug("existing policy fetched", "etag", existing.Etag)
		policy.Etag = existing.Etag
	}

	// Set the policy
	return s.updater.SetIamPolicy(ctx, policy)
}

func (s *setter) AddBinding(ctx context.Context, principalID, roleID string) (*models.HashicorpCloudResourcemanagerPolicy, error) {
	// Normalize the role
	role := normalizeRoleID(roleID)

	// Get the principal
	p, err := s.lookupPrincipal(principalID)
	if err != nil {
		return nil, fmt.Errorf("failed to look up principal %q: %w", principalID, err)
	}

	// Get the existing binding.
	s.logger.Debug("fetching existing policy")
	existing, err := s.updater.GetIamPolicy(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve existing policy: %w", err)
	}

	// Convert the policy to a map.
	bindings := ToMap(existing)

	// Check if the principal is already bound to the specified role
	if members, ok := bindings[role]; ok {
		if _, ok := members[principalID]; ok {
			return nil, fmt.Errorf("principal %q has existing role binding %q", principalID, role)
		}
	}

	members, ok := bindings[role]
	if !ok {
		members = make(map[string]*models.HashicorpCloudResourcemanagerPolicyBindingMemberType, 1)
		bindings[role] = members
	}

	members[p.MemberID] = p.MemberType
	s.logger.Debug("adding principal to policy", "principal", principalID, "principal_type", *p.MemberType, "role_id", role)
	return s.updater.SetIamPolicy(ctx, FromMap(existing.Etag, bindings))
}

func (s *setter) DeleteBinding(ctx context.Context, principalID, roleID string) (*models.HashicorpCloudResourcemanagerPolicy, error) {
	// Normalize the role
	role := normalizeRoleID(roleID)

	// Get the existing binding.
	s.logger.Debug("fetching existing policy in order to populate etag")
	existing, err := s.updater.GetIamPolicy(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve existing policy: %w", err)
	}

	// Find the binding to remove
	didDelete := false
	bindings := ToMap(existing)
	if members, ok := bindings[role]; ok {
		didDelete = true
		delete(members, principalID)
		if len(members) == 0 {
			delete(bindings, role)
		}
	}

	if !didDelete {
		s.logger.Debug("principal not found in policy", "principal", principalID, "role_id", role)
		return nil, fmt.Errorf("principal %q with role binding %q does not exist in policy", principalID, role)
	}

	s.logger.Debug("deleting principal from policy", "principal", principalID, "role_id", role)
	return s.updater.SetIamPolicy(ctx, FromMap(existing.Etag, bindings))
}

func (s *setter) lookupPrincipal(id string) (*models.HashicorpCloudResourcemanagerPolicyBindingMember, error) {
	s.logger.Debug("looking up principal", "id", id)
	params := iam_service.NewIamServiceBatchGetPrincipalsParams()
	params.OrganizationID = s.orgID
	params.View = (*string)(iamModels.HashicorpCloudIamPrincipalViewPRINCIPALVIEWBASIC.Pointer())
	params.PrincipalIds = []string{id}

	resp, err := s.iam.IamServiceBatchGetPrincipals(params, nil)
	if err != nil {
		return nil, err
	}

	if len(resp.Payload.Principals) != 1 {
		return nil, fmt.Errorf("failed to lookup principal %q", id)
	}

	p := resp.Payload.Principals[0]
	ptype, err := IamPrincipalTypeToBindingType(p)
	if err != nil {
		return nil, fmt.Errorf("failed to determine principal type: %w", err)
	}

	s.logger.Debug("discovered principal type", "id", id, "type", *ptype)
	return &models.HashicorpCloudResourcemanagerPolicyBindingMember{
		MemberID:   p.ID,
		MemberType: ptype,
	}, nil
}

func normalizeRoleID(role string) string {
	prefix := "roles/"
	if strings.HasPrefix(role, prefix) {
		return role
	}

	return fmt.Sprintf("%s%s", prefix, role)
}

// ToMap to map converts an IAM policy to a set of maps. The first map is keyed
// by Role ID, and the second map is keyed by PrincipalID.
func ToMap(p *models.HashicorpCloudResourcemanagerPolicy) map[string]map[string]*models.HashicorpCloudResourcemanagerPolicyBindingMemberType {
	bindings := make(map[string]map[string]*models.HashicorpCloudResourcemanagerPolicyBindingMemberType, len(p.Bindings))
	for _, b := range p.Bindings {
		bindings[b.RoleID] = make(map[string]*models.HashicorpCloudResourcemanagerPolicyBindingMemberType, len(b.Members))
		for _, m := range b.Members {
			bindings[b.RoleID][m.MemberID] = m.MemberType
		}
	}

	return bindings
}

// FromMap converts the map generated by ToMap to an IAM Policy object.
func FromMap(etag string, bindings map[string]map[string]*models.HashicorpCloudResourcemanagerPolicyBindingMemberType) *models.HashicorpCloudResourcemanagerPolicy {
	up := &models.HashicorpCloudResourcemanagerPolicy{
		Bindings: []*models.HashicorpCloudResourcemanagerPolicyBinding{},
		Etag:     etag,
	}

	for role, members := range bindings {
		b := &models.HashicorpCloudResourcemanagerPolicyBinding{
			Members: []*models.HashicorpCloudResourcemanagerPolicyBindingMember{},
			RoleID:  role,
		}

		for id, mtype := range members {
			m := &models.HashicorpCloudResourcemanagerPolicyBindingMember{
				MemberID:   id,
				MemberType: mtype,
			}

			b.Members = append(b.Members, m)
		}

		up.Bindings = append(up.Bindings, b)
	}

	return up
}

// IamPrincipalTypeToBindingType converts an IAM principal type to a resource
// manager binding member type.
func IamPrincipalTypeToBindingType(p *iamModels.HashicorpCloudIamPrincipal) (*models.HashicorpCloudResourcemanagerPolicyBindingMemberType, error) {
	if p == nil || p.Type == nil {
		return nil, fmt.Errorf("nil principal type")
	}

	switch *p.Type {
	case iamModels.HashicorpCloudIamPrincipalTypePRINCIPALTYPEUSER:
		return models.HashicorpCloudResourcemanagerPolicyBindingMemberTypeUSER.Pointer(), nil
	case iamModels.HashicorpCloudIamPrincipalTypePRINCIPALTYPEGROUP:
		return models.HashicorpCloudResourcemanagerPolicyBindingMemberTypeGROUP.Pointer(), nil
	case iamModels.HashicorpCloudIamPrincipalTypePRINCIPALTYPESERVICE:
		return models.HashicorpCloudResourcemanagerPolicyBindingMemberTypeSERVICEPRINCIPAL.Pointer(), nil
	default:
		return nil, fmt.Errorf("unsupported principal type (%s) for IAM Policy", *p.Type)
	}
}
