package logicmonitor

import (
	"fmt"
	"net/http"
	"context"
	"time"
	"os"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"

	"github.com/logicmonitor/lm-data-sdk-go/api/metrics"
	"github.com/logicmonitor/lm-data-sdk-go/model"

	"github.com/tidwall/gjson"
)

var sampleConfig string

type Logicmonitor struct {
	Apikey      string          `toml:"apikey"`
	URL         string          `toml:"url"`
	client *http.Client

	serializer serializers.Serializer
	Log telegraf.Logger `toml:"-"`
}

type Payload struct {
	Metrics   []Datapoint `json:"fields"`
	Datasource string `json:"name"`
	Tags []Tags `json:"tags"`
}

type Tags struct {
	Value string
}

type Datapoint struct {
	Name string
	Value string
}

func (l *Logicmonitor) SetSerializer(serializer serializers.Serializer) {
	l.serializer = serializer
}

func (*Logicmonitor) SampleConfig() string {
	return sampleConfig
}

func (l *Logicmonitor) Connect() error {
    // Make any connection required here
    return nil
}

func (l *Logicmonitor) Write(telegrafMetrics []telegraf.Metric) error {

	os.Setenv("LM_ACCOUNT", "lmjordikleriga")
	os.Setenv("LM_ACCESS_KEY", "E_3T(qcn)E9+4IJk)3B]x9)[[Cs4^e7K7K877C{Q")
	os.Setenv("LM_ACCESS_ID", "w2Np468gHqgXg8pPhpKH")

	var options []metrics.Option

	options = []metrics.Option{
		metrics.WithMetricBatchingEnabled(9 * time.Second),
	}

	lmMetric, err := metrics.NewLMMetricIngest(context.Background(), options...)
	if err != nil {
		fmt.Println("Error in initializing metric ingest :", err)
		return err
	}

	data, err := l.serializer.SerializeBatch(telegrafMetrics)
	
	jsonData := string(data)

	metricsArray := gjson.Get(jsonData,"metrics")

	fmt.Println("Sending Metrics")

    //iterate through all metrics
	metricsArray.ForEach(func(key, value gjson.Result) bool {

		fields := fmt.Sprint(value)

		rInput, dsInput, insInput := createResource(fields)

		dataPoints := gjson.Get(fields, "fields")

		dataPoints.ForEach(func(key, value gjson.Result) bool {

			dpInput := createMetric(fmt.Sprint(key), fmt.Sprint(value))

			err = lmMetric.SendMetrics(context.Background(), rInput, dsInput, insInput, dpInput)

			if err != nil {
				fmt.Println("Error in sending metric: ", err)
			}

			return true // keep iterating
		})

		return true // keep iterating
	})

	return nil
}

func createResource(json string) (model.ResourceInput, model.DatasourceInput, model.InstanceInput){
	
	var instName string
	if len(fmt.Sprint(gjson.Get(json, "tags.instance")))>0 {
		instName = (fmt.Sprint(gjson.Get(json, "tags.instance")))
	} else {
		instName = fmt.Sprint(gjson.Get(json, "name"))
	}

	dsName := ("Test_Telegraf_" + fmt.Sprint(gjson.Get(json, "name")))

	rInput := model.ResourceInput{
		ResourceName: fmt.Sprint(gjson.Get(json, "tags.host")),
		ResourceID:   map[string]string{"system.displayname": fmt.Sprint(gjson.Get(json, "tags.host"))},
		IsCreate:     true,
	}

	dsInput := model.DatasourceInput{
		DataSourceName: dsName,
		DataSourceGroup: "Telegraf",
	}

	insInput := model.InstanceInput{
		InstanceName: instName,
		InstanceProperties: map[string]string{"test": "datasdk"},
	}
	return rInput, dsInput, insInput
}

/*
func createResourceProps(json string) () {
	
}*/

func createMetric(name string, value string) (model.DataPointInput) {

	dpInput := model.DataPointInput{
		DataPointName:            name,
		Value:                    map[string]string{fmt.Sprintf("%d", time.Now().Unix()): fmt.Sprint(value)},
	}
	return dpInput
}

func (l *Logicmonitor) Close() error {
	return nil
}


func init() {
	outputs.Add("logicmonitor", func() telegraf.Output {
		return &Logicmonitor{}
	})
}
