package repeater

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/mark-by/proxy/internal/application"
	"github.com/mark-by/proxy/internal/domain/entity"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func scan(writer http.ResponseWriter, request *http.Request, app *application.App)  {
	vars := mux.Vars(request)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	r := app.Requests.Get(id)
	if r == nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	result := scanCmd(r)
	if len(result) == 0 {
		writer.WriteHeader(http.StatusNoContent)
		return
	}

	data, _ := json.Marshal(result)
	writer.Header().Set("Content-Type", "application/json")
	writer.Write(data)
}

func scanCmd(request *entity.Request) []string {
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	defer client.CloseIdleConnections()

	liveRequest, err := request.Revive()
	if err != nil {
		return nil
	}

	injections := getInjections()

	queryValues := liveRequest.URL.Query()
	var queryResults []string
	if len(queryValues) != 0 {
		scanValues(&scanOptions{
			Client:       &client,
			Request:      request,
			ParamsPlace:  "query",
			ValuesGetter: getQueryValues,
			Injections:   injections,
			Accum:        &queryResults,
		})
	}
	err = liveRequest.ParseForm()
	var formResults []string
	if err == nil && len(liveRequest.PostForm) != 0 {
		scanValues(&scanOptions{
			Client:       &client,
			Request:      request,
			ParamsPlace:  "form",
			ValuesGetter: getPostValues,
			Injections:   injections,
			Accum:        &formResults,
		})
	}

	return append(formResults, queryResults...)
}

type scanOptions struct {
	Client *http.Client
	Request *entity.Request
	ParamsPlace string
	ValuesGetter func(r *http.Request) url.Values
	Injections []string
	Accum *[]string
}

func readBody(response *http.Response) ([]byte, error) {
	var body io.ReadCloser
	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		body, _ = gzip.NewReader(response.Body)
	default:
		body = response.Body
	}

	data, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func scanValues(options *scanOptions) {
	r, _ := options.Request.Revive()
	for key, values := range options.ValuesGetter(r) {
		for _, value := range values {
			for _, injection := range options.Injections {
				logrus.Info("INJECTION: ", injection)
				logrus.Info("PARAMETER: ",  key)
				r, _ := options.Request.Revive()

				var injectedRequest *http.Request
				var err error
				if options.ParamsPlace == "form" {
					params := options.ValuesGetter(r)
					params.Set(key, value + injection)

					injectedRequest, err = http.NewRequest(r.Method, options.Request.URL, strings.NewReader(params.Encode()))
				} else {
					injectedRequest, err = http.NewRequest(r.Method, options.Request.URL, r.Body)
					options.ValuesGetter(injectedRequest).Set(key, value + injection)
				}

				entity.CopyHeaders(r, injectedRequest)

				response, err := options.Client.Do(injectedRequest)
				if err != nil {
					logrus.Error("fail send: ", err)
					continue
				}

				data, err := readBody(response)
				if err != nil {
					logrus.Error("Fail to read: ", err)
					continue
				}

				if strings.Contains(string(data), "root:") {
					logrus.Info("INJECTION FOUND IN ", options.ParamsPlace, " ", key, " parameter")
					logrus.Info("RESPONSE BODY: ", string(data))
					*options.Accum = append(*options.Accum, key)
				}
			}
		}
	}
}

func getQueryValues(r *http.Request) url.Values {
	return r.URL.Query()
}

func getPostValues(r *http.Request) url.Values {
	_ = r.ParseForm()
	return r.PostForm
}

func getInjections() []string {
	file, err := os.Open("cmd.injections")
	if err != nil {
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var injections []string
	for scanner.Scan() {
		injections = append(injections, scanner.Text())
	}

	return injections
}
