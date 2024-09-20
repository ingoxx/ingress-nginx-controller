package template_nginx

import (
	"bytes"
	"fmt"
	"k8s.io/klog/v2"
	"os"
	"text/template"
)

type RenderTemplate struct {
	GenerateName       string
	RenderTemplateName string
	MainTemplateName   string
}

func (rt *RenderTemplate) Render(data interface{}) error {
	// main template
	mainTmplStr, err := os.ReadFile(rt.MainTemplateName)
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("failed to parse %s template_nginx", rt.MainTemplateName))
		return err
	}

	mainTmpl, err := template.New("main").Parse(string(mainTmplStr))
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("error parsing template_nginx: %s data", rt.RenderTemplateName))
		return err
	}

	// render template
	renderTmplStr, err := os.ReadFile(rt.RenderTemplateName)
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("failed to parse %s template_nginx", rt.RenderTemplateName))
		return err
	}

	renderTmpl, err := template.New("render").Parse(string(renderTmplStr))
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("error parsing template_nginx: %s data", rt.RenderTemplateName))
		return err
	}

	var renderTpl bytes.Buffer
	if err := renderTmpl.Execute(&renderTpl, data); err != nil {
		return err
	}

	// insert the rendered template into the designated location block in the main template
	_, err = mainTmpl.New("servers").Parse(renderTpl.String())
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("error parsing template_nginx: %s data", rt.MainTemplateName))
		return err
	}

	var mainTpl bytes.Buffer
	if err = mainTmpl.Execute(&mainTpl, nil); err != nil {
		klog.ErrorS(err, fmt.Sprintf("rendering %s template_nginx failed", rt.RenderTemplateName))
		return err
	}

	if err := rt.Generate(mainTpl.Bytes()); err != nil {
		return err
	}

	return nil

}

func (rt *RenderTemplate) Generate(b []byte) error {
	conf := rt.GenerateName + "-test.conf"
	if err := os.WriteFile(conf, b, 0644); err != nil {
		klog.ErrorS(err, fmt.Sprintf("an error occurred while writing the generated content to %s", conf))
		return err
	}

	stat, err := os.Stat(conf)
	if err != nil || stat.Size() == 0 {
		klog.ErrorS(err, "detecting issues with the generated conf file")
		return err
	}

	return nil
}
