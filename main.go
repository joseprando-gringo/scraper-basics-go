package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func getReCaptchaV2Response(googleSiteKey string, pageURL string) string {
	values := map[string]interface{}{
		"type": "ReCaptchaV2",
		"args": map[string]interface{}{
			"googleSiteKey": googleSiteKey,
			"pageUrl":       pageURL,
		},
	}
	jsonData, err := json.Marshal(values)
	r, err := http.NewRequest("POST", "https://staging.gringo.com.vc/captcha-service-v2/v1/captcha", bytes.NewBuffer(jsonData))
	r.Header.Add("Authorization", "Api-Key 6c2bf5f6-545a-46c2-a8f3-4c5052d3a345")
	r.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	client.Timeout = time.Duration(time.Duration.Seconds(15))
	res, err := client.Do(r)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var data map[string]interface{}
	json.Unmarshal(bodyBytes, &data)
	return data["response"].(string)
}

func main() {
	renavam := "1076858306"
	placa := "FJQ8705"

	fmt.Println("1. Desabilitar SSL InsecureVerify")
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	fmt.Println("1. GET na pagina: https://www.ipva.fazenda.sp.gov.br/ipvanet_consulta/consulta.aspx")
	firstRequestResp, err := http.Get("https://www.ipva.fazenda.sp.gov.br/ipvanet_consulta/consulta.aspx")
	if err != nil {
		fmt.Printf("client: %s", err)
		os.Exit(1)
	}
	defer firstRequestResp.Body.Close()

	fmt.Println("3. Capturar o Cookie")
	cookies := firstRequestResp.Cookies()
	cookieString := cookies[0].Name + "=" + cookies[0].Value + "; " + cookies[1].Name + "=" + cookies[1].Value + "; "

	fmt.Println("4. Capturar os Input Hidden")
	doc, err := goquery.NewDocumentFromReader(firstRequestResp.Body)
	if err != nil {
		log.Fatal(err)
	}

	eventTarget, _ := doc.Find("#__EVENTTARGET").Attr("value")
	eventArgument, _ := doc.Find("#__EVENTARGUMENT").Attr("value")
	viewState, _ := doc.Find("#__VIEWSTATE").Attr("value")
	viewStateGenerator, _ := doc.Find("#__VIEWSTATEGENERATOR").Attr("value")
	eventValidation, _ := doc.Find("#__EVENTVALIDATION").Attr("value")

	inputRenavamName, _ := doc.Find("#conteudoPaginaPlaceHolder_txtRenavam").Attr("name")
	inputPlacaName, _ := doc.Find("#conteudoPaginaPlaceHolder_txtPlaca").Attr("name")
	buttonConsultarName, _ := doc.Find("#conteudoPaginaPlaceHolder_btn_Consultar").Attr("name")
	buttonConsultarValue, _ := doc.Find("#conteudoPaginaPlaceHolder_btn_Consultar").Attr("value")

	fmt.Println("5. Resolvendo Captcha")
	captchaResponse := getReCaptchaV2Response(
		"6Led7bcUAAAAAGqEoogy4d-S1jNlkuxheM7z2QWt",
		"https://www.ipva.fazenda.sp.gov.br/ipvanet_consulta/Consulta.aspx",
	)

	fmt.Println("6. Submeter o formulário (Placa, Renavam, Valor dos input hidden e Captcha): https://www.ipva.fazenda.sp.gov.br/ipvanet_consulta/Consulta.aspx")
	data := url.Values{}
	data.Set("__EVENTTARGET", eventTarget)
	data.Set("__EVENTARGUMENT", eventArgument)
	data.Set("__VIEWSTATE", viewState)
	data.Set("__VIEWSTATEGENERATOR", viewStateGenerator)
	data.Set("__EVENTVALIDATION", eventValidation)
	data.Set(inputRenavamName, renavam)
	data.Set(inputPlacaName, placa)
	data.Set("g-recaptcha-response", captchaResponse)
	data.Set(buttonConsultarName, buttonConsultarValue)

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, "https://www.ipva.fazenda.sp.gov.br/ipvanet_consulta/Consulta.aspx", strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Cookie", cookieString)
	r.Header.Add("Host", "www.ipva.fazenda.sp.gov.br")
	r.Header.Add("Origin", "https://www.ipva.fazenda.sp.gov.br")
	r.Header.Add("Referer", "https://www.ipva.fazenda.sp.gov.br/ipvanet_consulta/consulta.aspx")

	formPostResponse, _ := client.Do(r)
	fmt.Println(formPostResponse.Status)
	defer formPostResponse.Body.Close() // Se não estou usando, vale fechar?

	fmt.Println("7. GET na página aviso.aspx")
	r, _ = http.NewRequest(http.MethodGet, "https://www.ipva.fazenda.sp.gov.br/ipvanet_consulta/Pages/Aviso.aspx", nil)
	r.Header.Add("Cookie", cookieString)
	r.Header.Add("Host", "www.ipva.fazenda.sp.gov.br")
	r.Header.Add("Origin", "https://www.ipva.fazenda.sp.gov.br")
	r.Header.Add("Referer", "https://www.ipva.fazenda.sp.gov.br/ipvanet_consulta/Consulta.aspx")

	getDataPageResponse, err := client.Do(r)
	if err != nil {
		panic(err)
	}
	defer getDataPageResponse.Body.Close()

	fmt.Println("8. Extrair os elementos da página de débitos")
	doc, _ = goquery.NewDocumentFromReader(getDataPageResponse.Body)

	marcaModelo := doc.Find("#conteudoPaginaPlaceHolder_txtMarcaModelo").Text()
	fmt.Println(marcaModelo)

	// fmt.Println("9. Validar se existem multas (detalhamento)")
	// fmt.Println("7.1. Multas encontradas")
}
