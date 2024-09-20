package config

const (
	ConfDir        = "/etc/nginx/conf.d"
	NginxTmpl      = "/rootfs/etc/nginx/template/nginx.tmpl"
	ServerTmpl     = "/rootfs/etc/nginx/template/server.tmpl"
	MainServerTmpl = "/rootfs/etc/nginx/template/mainServer.tmpl"
	DefaultTmpl    = "/rootfs/etc/nginx/template/defaultBackend.tmpl"
	SslPath        = "/etc/nginx/ssl"
	TlsCrt         = "tls.crt"
	TlsKey         = "tls.key"
	Pid            = "/var/run/nginx.pid"
	Bin            = "/usr/sbin/nginx"
	MainConf       = "/etc/nginx/nginx.conf"
)
