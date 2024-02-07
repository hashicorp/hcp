package agent

/*
type TFEModuleDetailsOperation struct {
	TFCToken string

	Namespace string
	Name      string
	Provider  string
	Version   string
}

func (s *TFEModuleDetailsOperation) Run(ctx context.Context, log hclog.Logger) error {
	moduleSource := fmt.Sprintf("%s/%s/%s/%s", s.Namespace, s.Name, s.Provider, s.Version)
	modulePath := fmt.Sprintf("https://app.terraform.io/api/registry/v1/modules/" + moduleSource)

	req, err := http.NewRequest("GET", modulePath, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to get module details for module: "+moduleSource)
	}
	req = req.WithContext(ctx)
	req.Header.Set("Authorization", "Bearer "+s.TFCToken)
	req.Header.Set("content-type", "application/vnd.api+json")

	client := &http.Client{}
	rawResp, err := client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "failed to get module details for module: "+moduleSource)
	}
	defer rawResp.Body.Close()
	resp := struct {
		Root struct {
			Readme string `json:"readme"`
		}
	}{}
	err = json.NewDecoder(rawResp.Body).Decode(&resp)
	if err != nil {
		return errors.Wrapf(err, "failed to get module details for module: "+moduleSource)
	}
	return &pb.TFModuleDetails{Readme: resp.Root.Readme}, nil
}
*/
