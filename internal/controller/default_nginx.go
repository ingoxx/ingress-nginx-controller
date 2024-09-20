package controller

import (
	"fmt"
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/nginx"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/pkg/utils/template_nginx"
)

type ConfHandler struct {
}

func NewConfHandler() ConfHandler {
	return ConfHandler{}
}

func (c ConfHandler) UpdateDefaultConf(parser *template_nginx.RenderTemplate) error {
	var servers = new(ingressv1.Server)
	var cfg = struct {
		Server *ingressv1.Server
	}{
		Server: servers,
	}

	fmt.Println("UpdateDefaultConf >>> ", cfg)

	if err := parser.Render(cfg); err != nil {
		fmt.Println("Render >>> ", err)
		return err
	}

	if err := nginx.Reload(parser.GenerateName); err != nil {
		fmt.Println("UpdateDefaultConf >>> ", err)
		return err
	}

	return nil
}
