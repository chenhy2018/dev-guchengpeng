package configs

type Service struct {
	Base

	AuthInfo *AuthInfo `inject`
}

type batchConfigInput struct {
	Files []struct {
		Name    string `json:"name"`
		Version int    `json:"version"`
	} `json:"files"`
}

type batchConfigOutput struct {
	Files []batchConfigFile `json:"files"`
}

type batchConfigFile struct {
	Code    int      `json:"code"`
	Message string   `json:"message,omitempty"`
	Data    repoFile `json:"data,omitempty"`
}

func (p *Service) BatchConfig() {
	app := p.AuthInfo.Service

	var input batchConfigInput
	err := p.Params.BindJsonBody(&input, true)
	if err != nil {
		p.Log.Error("BindJsonBody", err)
		p.Rw.WriteHeader(500)
		return
	}

	if len(input.Files) == 0 {
		p.Rw.WriteHeader(400)
		return
	}

	files := make([]batchConfigFile, len(input.Files))
	for i, req := range input.Files {
		file, err := Repo.LoadFile(app, req.Name, req.Version)
		if err == ErrNotExists {
			files[i].Code = 404
			files[i].Data.Name = req.Name
			continue
		}
		if err != nil {
			files[i].Code = 500
			files[i].Message = err.Error()
			files[i].Data.Name = req.Name
			continue
		}
		files[i].Code = 200
		files[i].Data.Name = file.Name
		files[i].Data.Version = file.Version
		files[i].Data.Content = file.Content
	}

	output := batchConfigOutput{files}
	err = p.Render.JSON(&output)
	if err != nil {
		p.Log.Error("json.Marshal", err)
		p.Rw.WriteHeader(500)
		return
	}
}
