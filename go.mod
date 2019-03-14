module k8s.io/helm

go 1.12

require (
	cloud.google.com/go v0.34.0
	contrib.go.opencensus.io/exporter/ocagent v0.2.0
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78
	github.com/Azure/go-autorest v11.2.8+incompatible
	github.com/BurntSushi/toml v0.3.1
	github.com/MakeNowJust/heredoc v0.0.0-20171113091838-e9091a26100e
	github.com/Masterminds/semver v1.4.2
	github.com/Masterminds/sprig v2.16.0+incompatible
	github.com/Masterminds/vcs v1.13.0
	github.com/PuerkitoBio/purell v1.1.0
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578
	github.com/Shopify/logrus-bugsnag v0.0.0-20171204204709-577dee27f20d
	github.com/aokoli/goutils v1.0.1
	github.com/asaskevich/govalidator v0.0.0-20180315120708-ccb8e960c48f
	github.com/beorn7/perks v0.0.0-20180321164747-3a771d992973
	github.com/bshuster-repo/logrus-logstash-hook v0.4.1
	github.com/bugsnag/bugsnag-go v1.3.2
	github.com/bugsnag/panicwrap v1.2.0
	github.com/census-instrumentation/opencensus-proto v0.1.0
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1
	github.com/containerd/containerd v1.2.1
	github.com/cpuguy83/go-md2man v1.0.8
	github.com/davecgh/go-spew v1.1.1
	github.com/deislabs/oras v0.3.3
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/distribution v2.7.0+incompatible
	github.com/docker/docker v0.0.0-20181221150755-2cb26cfe9cbf
	github.com/docker/go-metrics v0.0.0-20181218153428-b84716841b82
	github.com/docker/go-units v0.3.3
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c
	github.com/emicklei/go-restful v2.8.0+incompatible
	github.com/evanphx/json-patch v3.0.0+incompatible
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d
	github.com/fatih/camelcase v1.0.0
	github.com/garyburd/redigo v1.6.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-openapi/jsonpointer v0.17.2
	github.com/go-openapi/jsonreference v0.17.2
	github.com/go-openapi/spec v0.17.2
	github.com/go-openapi/swag v0.17.2
	github.com/gobwas/glob v0.2.3
	github.com/gogo/protobuf v1.2.0
	github.com/golang/groupcache v0.0.0-20181024230925-c65c006176ff
	github.com/golang/protobuf v1.2.0
	github.com/google/btree v0.0.0-20180813153112-4030bb1f1f0c
	github.com/google/gofuzz v0.0.0-20170612174753-24818f796faf
	github.com/google/uuid v1.1.0
	github.com/googleapis/gnostic v0.2.0
	github.com/gophercloud/gophercloud v0.0.0-20181221023737-94924357ebf6
	github.com/gorilla/handlers v1.4.0
	github.com/gorilla/mux v1.7.0
	github.com/gosuri/uitable v0.0.0-20160404203958-36ee7e946282
	github.com/gregjones/httpcache v0.0.0-20181110185634-c63ab54fda8f
	github.com/hashicorp/golang-lru v0.5.0
	github.com/huandu/xstrings v1.2.0
	github.com/imdario/mergo v0.3.6
	github.com/inconshreveable/mousetrap v1.0.0
	github.com/json-iterator/go v1.1.5
	github.com/kardianos/osext v0.0.0-20170510131534-ae77be60afb1
	github.com/konsorten/go-windows-terminal-sequences v1.0.1
	github.com/mailru/easyjson v0.0.0-20180823135443-60711f1a8329
	github.com/mattn/go-runewidth v0.0.4
	github.com/mattn/go-shellwords v1.0.3
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/miekg/dns v0.0.0-20181005163659-0d29b283ac0f
	github.com/mitchellh/go-wordwrap v1.0.0
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd
	github.com/modern-go/reflect2 v0.0.0-20180701023420-4b7aa43c6742
	github.com/opencontainers/go-digest v1.0.0-rc1
	github.com/opencontainers/image-spec v1.0.1
	github.com/petar/GoLLRB v0.0.0-20130427215148-53be0d36a84c
	github.com/peterbourgon/diskv v2.0.1+incompatible
	github.com/pkg/errors v0.8.0
	github.com/pmezard/go-difflib v1.0.0
	github.com/prometheus/client_golang v0.9.2
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90
	github.com/prometheus/common v0.2.0
	github.com/prometheus/procfs v0.0.0-20190129233650-316cf8ccfec5
	github.com/russross/blackfriday v1.5.2
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.3.0
	github.com/xenolf/lego v0.0.0-20160613233155-a9d8cec0e656
	github.com/yvasiyarov/go-metrics v0.0.0-20150112132944-c25f46c4b940
	github.com/yvasiyarov/gorelic v0.0.6
	github.com/yvasiyarov/newrelic_platform_go v0.0.0-20140908184405-b21fdbd4370f
	go.opencensus.io v0.18.0
	golang.org/x/crypto v0.0.0-20181203042331-505ab145d0a9
	golang.org/x/net v0.0.0-20181220203305-927f97764cc3
	golang.org/x/oauth2 v0.0.0-20181203162652-d668ce993890
	golang.org/x/sync v0.0.0-20181221193216-37e7f081c4d4
	golang.org/x/sys v0.0.0-20181221143128-b4a75ba826a6
	golang.org/x/text v0.3.0
	golang.org/x/time v0.0.0-20181108054448-85acf8d2951c
	google.golang.org/api v0.0.0-20181221000618-65a46cafb132
	google.golang.org/appengine v1.4.0
	google.golang.org/genproto v0.0.0-20181221175505-bd9b4fb69e2f
	google.golang.org/grpc v1.17.0
	gopkg.in/inf.v0 v0.9.1
	gopkg.in/square/go-jose.v1 v1.1.2
	gopkg.in/square/go-jose.v2 v2.3.0
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0-20190222213804-5cb15d344471
	k8s.io/apiextensions-apiserver v0.0.0-20190221221350-bfb440be4b87
	k8s.io/apimachinery v0.0.0-20190221213512-86fb29eff628
	k8s.io/apiserver v0.0.0-20190221215341-5838f549963b
	k8s.io/cli-runtime v0.0.0-20181221202950-8abb1aeb8307
	k8s.io/client-go v0.0.0-20190228174230-b40b2a5939e4
	k8s.io/klog v0.1.0
	k8s.io/kube-openapi v0.0.0-20181114233023-0317810137be
	k8s.io/kubernetes v0.0.0-20190305150815-6c1e64b94a3e
	k8s.io/utils v0.0.0-20181221173059-8a16e7dd8fb6
	sigs.k8s.io/kustomize v1.0.11
	sigs.k8s.io/yaml v1.1.0
	vbom.ml/util v0.0.0-20180919145318-efcd4e0f9787
)
