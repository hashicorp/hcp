package format

import (
	"encoding/json"
	"fmt"
	"testing"

	example "github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"
	"github.com/hashicorp/hcp/internal/pkg/iostreams"
	"github.com/stretchr/testify/require"
)

// TODO move this test elsewhere. This is just for debugging purposes.
func TestTemp(t *testing.T) {
	j := `{"action_configs":[{"id":"00000000-0000-0000-0000-000000000000", "action_url":"", "name":"Agent Smith", "request":{"agent":{"op":{"id":"Agent Smith", "body":"", "action_run_id":"", "group":"Enforcements"}}}, "description":"test description", "created_at":"2024-08-16T18:11:19.777071Z"}, {"id":"11111111-1111-1111-1111-111111111111", "action_url":"", "name":"Example", "request":{"custom":{"method":"GET", "headers":[], "url":"https://hashicorp.com", "body":""}}, "description":"Runs an action against https://hashicorp.com", "created_at":"2024-06-13T17:31:17.436255Z"}, {"id":"22222222-2222-2222-2222-222222222222", "action_url":"", "name":"Variables", "request":{"custom":{"method":"GET", "headers":[], "url":"https://${var.company}.com", "body":""}}, "description":"An action to test the variables feature.", "created_at":"2024-08-07T21:56:00.043627Z"}], "pagination":{"next_page_token":"", "previous_page_token":""}}`
	thing := example.HashicorpCloudWaypointListActionConfigResponse{}
	err := json.Unmarshal([]byte(j), &thing)
	require.NoError(t, err)

	// dsp := DisplayFields(thing.ActionConfigs, Pretty)
	// tmpl := prettyPrintTemplate(dsp)
	// fmt.Println(tmpl)

	io := iostreams.Test()
	out := New(io)
	err = out.Show(thing.ActionConfigs, Pretty)

	require.NoError(t, err)
	fmt.Println(io.Output.String())
	t.FailNow()
}
