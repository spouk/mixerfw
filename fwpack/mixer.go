//---------------------------------------------------------------------------
//  Copyright(c)2017 by Spouk
//  https://github.com/spouk
//  https://www.spouk.ru
// --------------------------------------------------------------------------
//  Простой фреймворк с поддержкой контекста на мультиплекcоре
//  httprouter - 'https://github.com/julienschmidt/httprouter'
//---------------------------------------------------------------------------
package Mixer

import (
	"sync"
	"mixerfw/fwpack/ext"
	"net/http"
	"runtime"
	"reflect"
	"os"
	"io"
	"encoding/json"
	"time"
	"strings"
	"net/url"
	"context"
	"fmt"
)

//---------------------------------------------------------------------------
//  TYPE: определение типов и интерфейсов
//---------------------------------------------------------------------------
type (
	//---------------------------------------------------------------------------
	//  MIXER: фреймворк
	//---------------------------------------------------------------------------
	Mixer struct {
		router    httprouter.Router
		pool      sync.Pool
		server    *http.Server
		tlsserver *http.Server
		//миддлы
		middlestock map[string][]MixerMiddleware
		//карта рутеров для визуализации и просмотра
		stockMapRoute []MixerMapRoute
		//логгер
		logger MixerLogger
		//рендер шаблонов
		render MixerRender
		//базовый конфиг
		config MixerConfigDefault
		//статичный контейнер
		staticData *MixerStaticData
	}
	//---------------------------------------------------------------------------
	//  MIXERSTATICDATA
	//---------------------------------------------------------------------------
	//cook, key, value
	MixerStaticData struct {
		Data map[string]map[string]string
	}
	//---------------------------------------------------------------------------
	//  MIXERCONFIG: базовый конфиг
	//---------------------------------------------------------------------------
	MixerConfigDefault struct {
		//путь к шаблонам
		TemplatePath string
		//отладка шаблонов
		TemplateDebug bool
		//адрес сервера
		Address    string
		AdressHTTP string
		//таймауты сервера на чтение/запись
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
		//TLS сертификат и ключ (полные пути надо указывать)
		CertFile string
		KeyFile  string

		RedirectTrailingSlash bool
		RedirectFixedPath     bool
	}
	//---------------------------------------------------------------------------
	//  MIXERMIDDLEWARE: тип миддла фреймворка
	//---------------------------------------------------------------------------
	MixerMiddleware func(handler MixerHandler) MixerHandler
	//---------------------------------------------------------------------------
	//  MIXERHANDLER: тип обработчика фреймворка
	//---------------------------------------------------------------------------
	MixerHandler func(carry *MixerCarry) error
	//---------------------------------------------------------------------------
	//  MIXERCARRY: тип несущей фреймворка в конечных обработчиках
	//---------------------------------------------------------------------------
	MixerCarry struct {
		w http.ResponseWriter
		r *http.Request
		p httprouter.Params
		m *Mixer
	}
	//---------------------------------------------------------------------------
	//  MIXERMAPROUTE: тип стека карты рутеров для визуализации фреймворка
	//---------------------------------------------------------------------------
	MixerMapRoute struct {
		Path     string
		Method   string
		Prefix   string
		Handler  string
		HandlerX string
	}
	//---------------------------------------------------------------------------
	//  MIXERSUBDOMAIN: субдомены
	//---------------------------------------------------------------------------
	MixerSubdomain struct {
		m      *Mixer
		prefix string
	}
	//---------------------------------------------------------------------------
	//  MIXERLOGGING: логгер (интерфейс)
	//---------------------------------------------------------------------------
	MixerLogger interface {
		FPrintf(out io.Writer, format string, v ...interface{})
		Printf(format string, v ...interface{})
		Info(message string)
		Error(message string)
		Warning(message string)
		Fatal(v interface{})
	}
	//---------------------------------------------------------------------------
	//  MIXERRENDER: рендер шаблонов ( интерфейс )
	//---------------------------------------------------------------------------
	MixerRender interface {
		Render(name string, data interface{}, resp io.Writer) (err error)
		RenderCode(httpCode int, name string, data interface{}, w io.Writer) (err error)
		AddUserFilter(name string, f interface{})
		AddFilters(stack map[string]interface{})
		ReloadTemplate()
		RenderTxt(httpCode int, name string, w io.Writer) (err error)
	}
)

//---------------------------------------------------------------------------
//  VARs: переменные
//---------------------------------------------------------------------------
var (
	staticMux = http.NewServeMux()
)

//---------------------------------------------------------------------------
//  MIXER: функционал
//---------------------------------------------------------------------------
func NewMixer(config MixerConfigDefault) *Mixer {
	m := &Mixer{
		pool: sync.Pool{},
	}
	//создаю пул с дефолтной функцией
	m.pool.New = func() interface{} {
		return &MixerCarry{
			m: m,
			r: &http.Request{},
		}

	}
	//создаю mixerStaticData
	m.staticData = m.newStaticMixerData()
	//создаю дефолтный ассоциативный массив для миддов
	m.middlestock = make(map[string][]MixerMiddleware)
	//создаю дефолтный логгер
	m.logger = NewMixerLogger(MIXERPREFIX, os.Stdout)
	//создаю дефолтный рендер
	m.render = NewMixerRenderDefault(config.TemplatePath, config.TemplateDebug, m.logger)
	//config
	m.config = config

	m.router.RedirectTrailingSlash = config.RedirectTrailingSlash
	m.router.RedirectFixedPath = config.RedirectFixedPath
	return m

}
func (m *Mixer) ExportMixerLogger() MixerLogger {
	return m.logger
}
func (m *Mixer) SetRender(render MixerRender) {
	m.render = render
}
func (m *Mixer) SetLogger(logger MixerLogger) {
	m.logger = logger
}
func (m *Mixer) Run() {
	//обновление шаблонов
	m.render.ReloadTemplate()
	//создание сервер инстанса
	m.server = &http.Server{
		Addr:         m.config.Address,
		Handler:      m,
		ReadTimeout:  m.config.ReadTimeout,
		WriteTimeout: m.config.WriteTimeout,
	}
	m.logger.Printf(MSG_STARTSERVER, m.config.Address)
	m.logger.Fatal(m.server.ListenAndServe())
}
func (m *Mixer) RunTLS() {
	//обновление шаблонов
	//создание сервер инстанса
	m.tlsserver = &http.Server{
		Addr:         m.config.AdressHTTP,
		Handler:      m,
		ReadTimeout:  m.config.ReadTimeout,
		WriteTimeout: m.config.WriteTimeout,
	}
	m.logger.Printf(MSG_STARTSERVER, m.config.AdressHTTP)
	m.render.ReloadTemplate()
	//cert := "/home/spouk/go/src/spouk.ru/ssl/fullchain1.pem"
	//key := "/home/spouk/go/src/spouk.ru/ssl/privkey1.pem"
	//err:=m.tlsserver.ListenAndServeTLS(cert, key)
	d, e := os.Getwd()
	m.logger.Printf("GETWD: %s:%s\n", d, e)
	err := m.tlsserver.ListenAndServeTLS(m.config.CertFile, m.config.KeyFile)
	m.logger.Printf("CET+KEY: %s:%s\n", m.config.CertFile, m.config.KeyFile)
	if err != nil {
		m.logger.Printf("[error][startTLSHTTPSServer] `%v`\n", err)
		m.logger.Fatal(err)
	}
	m.logger.Printf("[ok-start][startTLSHTTPSServer] `%s`\n", m.config.AdressHTTP)

}
func (m *Mixer) RunTLSTEST(addr, cert, key string) {
	//обновление шаблонов
	//создание сервер инстанса
	m.tlsserver = &http.Server{
		Addr:         addr,
		Handler:      m,
		ReadTimeout:  m.config.ReadTimeout,
		WriteTimeout: m.config.WriteTimeout,
	}
	m.logger.Printf(MSG_STARTSERVER, m.config.AdressHTTP)
	m.render.ReloadTemplate()
	m.logger.Fatal(m.tlsserver.ListenAndServeTLS(cert, key))

}

//---------------------------------------------------------------------------
//  MIXER: работа с миддлами
//---------------------------------------------------------------------------
func (m *Mixer) AddMiddleware(prefix string, middleware MixerMiddleware) {
	//добавляет в общий сток новый миддл
	//миддлы идут по порядку и обрабатываются соответственно тоже по порядку добавления в стек
	if prefix == "" {
		m.middlestock[""] = append(m.middlestock[""], middleware)
	} else {
		m.middlestock[prefix] = append(m.middlestock[prefix], middleware)
	}
}
func (m *Mixer) wrapperMiddlewares(prefix string, handler MixerHandler) MixerHandler {
	//враппер миддлами для обработчиков, что берутся из пула
	//обертываем с конца стека
	stock := m.middlestock[prefix]
	for x := len(stock) - 1; x >= 0; x-- {
		handler = stock[x](handler)
	}
	return handler
}

//---------------------------------------------------------------------------
//  MIXER: работа с пулом
//---------------------------------------------------------------------------
func (m *Mixer) getPool(w http.ResponseWriter, r *http.Request) *MixerCarry {
	newcarry := m.pool.Get().(*MixerCarry)
	newcarry.w = w
	newcarry.r = r.WithContext(r.Context())
	return newcarry
}
func (m *Mixer) PutPool(r *MixerCarry) {
	m.pool.Put(r)
}

// REALPATH: должна начинаться с `.` (точки)
//example: chttp.Handle("/hacker/", http.StripPrefix("/hacker/", http.FileServer(http.Dir("./hacker"))))
//example: chttp.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./static"))))
//example:  StaticAddPrefix("/", "./")
func (m *Mixer) StaticAddPrefix(prefix string, realpath string, ) {
	staticMux.Handle(prefix, http.StripPrefix(prefix, http.FileServer(http.Dir(realpath))))
}
func (m *Mixer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	//добавляю проверку для корневого статичного контента
	if strings.Contains(r.URL.Path, ".") {
		//m.logger.Printf("FOUND STATIC PATH `%s`\n", r.URL.Path)
		staticMux.ServeHTTP(w, r)
	} else {
		//общоая обработка `httprouter`
		m.router.ServeHTTP(w, r)
	}
}

//---------------------------------------------------------------------------
//  MIXER: добавление рутеров + обработчиков
//---------------------------------------------------------------------------
func (m *Mixer) convertMixerHandlertohttprouterHandler(handler MixerHandler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newmux := m.getPool(w, r)
		handler(newmux)
		m.PutPool(newmux)
	})

}
func (m *Mixer) addRoute(method, path, prefix string, s MixerHandler) {
	//получаем имя хэндлера и добавляем в карту рутеров для визуализации
	nameHandler := runtime.FuncForPC(reflect.ValueOf(s).Pointer()).Name()
	m.stockMapRoute = append(m.stockMapRoute, MixerMapRoute{
		Path:    path,
		Method:  method,
		Prefix:  prefix,
		Handler: nameHandler,
	})
	//обертываем хэндлер миддлами из стека
	hu := m.wrapperMiddlewares(prefix, s)
	m.router.Handle(method, prefix+path, httprouter.Handle(func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		newmux := m.getPool(w, r)
		newmux.p = ps
		hu(newmux)
		m.PutPool(newmux)
	}))
}
func (m *Mixer) Multi(methods []string, path string, s MixerHandler) {
	//добавление хэндлера с несколькими видами http методов
	fu := func(method string) bool {
		for _, met := range ACCEPTHTTPMETHOD {
			if met == method {
				return true
			}
		}
		return false
	}
	for _, methodChecked := range methods {
		if fu(methodChecked) {
			m.addRoute(methodChecked, path, "", s)
		} else {
			m.logger.Error(ERROR_HTTPMETHODNOTACCEPT)
		}
	}
}
func (m *Mixer) HTTP_GET(path string, s MixerHandler) {
	m.addRoute("GET", path, "", s)
}
func (m *Mixer) HTTP_POST(path string, s MixerHandler) {
	m.addRoute("POST", path, "", s)
}
func (m *Mixer) HTTP_DELETE(path string, s MixerHandler) {
	m.addRoute("DELETE", path, "", s)
}
func (m *Mixer) HTTP_UPDATE(path string, s MixerHandler) {
	m.addRoute("UPDATE", path, "", s)
}

//---------------------------------------------------------------------------
// MIXER: информационный функционал
//---------------------------------------------------------------------------
func (m *Mixer) ShowRoutingMap() {
	for _, x := range m.stockMapRoute {
		fmt.Fprintf(os.Stdout, MIXERPREFIX+"[%7s] `%10s` `%20s`  `%+v`   `%v`\n", x.Prefix, x.Method, x.Path, x.Handler, x.HandlerX)
	}
}
func (m *Mixer) RoutingMapGet() []MixerMapRoute {
	return m.stockMapRoute
}
func (m *Mixer) ShowMiddlewares() {
	m.logger.Printf("Middlewares: %v\n", m.middlestock)
}

//---------------------------------------------------------------------------
//  MIXER: субдомайны
//---------------------------------------------------------------------------
func (m *Mixer) Subdomain(prefix string) *MixerSubdomain {
	return &MixerSubdomain{m: m, prefix: prefix}

}
func (s *MixerSubdomain) AddMiddleware(middleware MixerMiddleware) {
	s.m.AddMiddleware(s.prefix, middleware)
}
func (s *MixerSubdomain) Multi(methods []string, path string, hand MixerHandler) {
	//добавление хэндлера с несколькими видами http методов
	fu := func(method string) bool {
		for _, met := range ACCEPTHTTPMETHOD {
			if met == method {
				return true
			}
		}
		return false
	}
	for _, methodChecked := range methods {
		if fu(methodChecked) {
			s.m.addRoute(methodChecked, path, s.prefix, hand)
		} else {
			s.m.logger.Error(ERROR_HTTPMETHODNOTACCEPT)
		}
	}
}
func (s *MixerSubdomain) HTTP_GET(path string, h MixerHandler) {
	s.m.addRoute("GET", path, s.prefix, h)
}
func (s *MixerSubdomain) HTTP_POST(path string, h MixerHandler) {
	s.m.addRoute("POST", path, s.prefix, h)
}
func (s *MixerSubdomain) HTTP_DELETE(path string, h MixerHandler) {
	s.m.addRoute("DELETE", path, s.prefix, h)
}
func (s *MixerSubdomain) HTTP_UPDATE(path string, h MixerHandler) {
	s.m.addRoute("UPDATE", path, s.prefix, h)
}

//---------------------------------------------------------------------------
//  MIXER: статичные хэндлеры + определение 404 и 405
//---------------------------------------------------------------------------
func (m *Mixer) StaticFiles(realpath, wwwpath string) {
	m.router.ServeFiles(realpath, http.Dir(wwwpath))
}

//func (m *Mixer) StaticSingleFile()
func (m *Mixer) Error404Handler(handler MixerHandler) {
	m.router.NotFound = m.convertMixerHandlertohttprouterHandler(handler)
}
func (m *Mixer) Error405Handler(handler MixerHandler) {
	m.router.MethodNotAllowed = m.convertMixerHandlertohttprouterHandler(handler)
}

//---------------------------------------------------------------------------
//  MIXER: врапперы для интерфейса render (добавление новых функций)
//---------------------------------------------------------------------------
func (m *Mixer) AddRenderNewTemplatesFunction(mnemonic string, f interface{}) {
	m.render.AddUserFilter(mnemonic, f)
}
func (m *Mixer) AddRenderNewTemplatesFunctionList(stack map[string]interface{}) {
	m.render.AddFilters(stack)
}
//---------------------------------------------------------------------------
//  MixerStaticData
//---------------------------------------------------------------------------
func (m *Mixer) newStaticMixerData() *MixerStaticData {
	n := new(MixerStaticData)
	n.Data = make(map[string]map[string]string)
	return n
}
func (m *MixerStaticData) add(cook , key , value  string) bool{
	m.Data[cook][key] = value
	return true
}
func (m *MixerStaticData) get(cook , key string) (string, bool){
	if data_cook, found := m.Data[cook]; found {
		return data_cook[key], true
	} else {
		return "", false
	}
}
func (m *Mixer) StaticDataAdd(cook , key , value  string) bool{
	return m.staticData.add(cook, key, value)
}
func (m *Mixer) StaticDataGet(cook , key  string) (string, bool){
	return m.staticData.get(cook, key)
}
//---------------------------------------------------------------------------
//  MIXERCARRY: функционал
//---------------------------------------------------------------------------
func (m *MixerCarry) InStringMap(v string, mu []string) bool {
	for _, x := range mu {
		if v == x {
			return true
		}
	}
	return false
}
//---------------------------------------------------------------------------
//	путь с названиями аргументов
//---------------------------------------------------------------------------
func (m *MixerCarry) RealPath() string {

	_, result, _, _ := m.m.router.LookupRoute(m.r.Method, m.r.URL.Path)
	return result
}
//---------------------------------------------------------------------------
//рендеринг шаблона
//---------------------------------------------------------------------------
func (m *MixerCarry) Render(name string, data interface{}) error {
	return m.m.render.Render(name, data, m.w)
}
//---------------------------------------------------------------------------
//перегузка дерева шаблонов
//---------------------------------------------------------------------------
func (m *MixerCarry) ReloadTemplates() {
	m.m.render.ReloadTemplate()
	return
}
//---------------------------------------------------------------------------
//  рендеринг текстового файла
//---------------------------------------------------------------------------
func (m *MixerCarry) RenderTxt(httpcode int, name string) error {
	return m.m.render.RenderTxt(httpcode, name, m.w)
}
//---------------------------------------------------------------------------
//  рендеринг
//---------------------------------------------------------------------------
func (m *MixerCarry) RenderCode(httpcode int, name string, data interface{}) error {
	return m.m.render.RenderCode(httpcode, name, data, m.w)
}
//---------------------------------------------------------------------------
//  возвращает текущий инстанс-запроса
//---------------------------------------------------------------------------
func (m *MixerCarry) Request() *http.Request {
	return m.r
}
//---------------------------------------------------------------------------
//  возвращает текущий инстанс-ответа
//---------------------------------------------------------------------------
func (m *MixerCarry) Response() http.ResponseWriter {
	return m.w
}
//---------------------------------------------------------------------------
//  возвращает текущие параметры
//---------------------------------------------------------------------------
func (m *MixerCarry) Params() httprouter.Params {
	return m.p
}
//---------------------------------------------------------------------------
//  редирект на указанный путь
//---------------------------------------------------------------------------
func (m *MixerCarry) Redirect(path string) error {
	http.Redirect(m.w, m.r, path, http.StatusFound)
	return nil
}
//---------------------------------------------------------------------------
//  возвращает referer - последний http адрес с которого пришел request
//---------------------------------------------------------------------------
func (m *MixerCarry) Referer() *url.URL {
	res, _ := url.Parse(m.r.Referer())
	return res
}
//---------------------------------------------------------------------------
//  записывает html в responseWriter
//---------------------------------------------------------------------------
func (m *MixerCarry) WriteHTML(httpcode int, text string) (error) {
	resp := m.w
	resp.Header().Set(ContentType, TextHTMLCharsetUTF8)
	resp.WriteHeader(httpcode)
	resp.Write([]byte(text))
	return nil
}
//---------------------------------------------------------------------------
//  записывает json(byte format) в responseWriter
//---------------------------------------------------------------------------
func (m *MixerCarry) JSONB(httpcode int, b []byte) (error) {
	resp := m.w
	resp.Header().Set(ContentType, ApplicationJavaScriptCharsetUTF8)
	resp.WriteHeader(httpcode)
	resp.Write(b)
	return nil
}
//---------------------------------------------------------------------------
//  записывает json в responseWriter
//---------------------------------------------------------------------------
func (m *MixerCarry) JSON(code int, answer interface{}) (err error) {
	b, err := json.Marshal(answer)
	if err != nil {
		return err
	}
	return m.JSONB(code, b)
}
//---------------------------------------------------------------------------
//  добавляет новый элемент в контекст текущего запроса
//---------------------------------------------------------------------------
func (m *MixerCarry) Set(key string, value interface{}) {
	ctx := context.WithValue(m.r.Context(), key, value)
	m.r = m.r.WithContext(ctx)
}
//---------------------------------------------------------------------------
//  получает элемент из текущего контекста
//---------------------------------------------------------------------------
func (m *MixerCarry) Get(key string) (value interface{}) {
	return m.r.Context().Value(key)
}
//---------------------------------------------------------------------------
//  возвращает текущий путь
//---------------------------------------------------------------------------
func (m *MixerCarry) Path() (value string) {
	return m.r.URL.Path
}
//---------------------------------------------------------------------------
//  возвращает параметр из request->path
//---------------------------------------------------------------------------
func (m *MixerCarry) GetParam(key string) string {
	return m.p.ByName(key)
}
//---------------------------------------------------------------------------
//  возвращает все элементы запроса  url (&={value}&={value1} etc...)
//---------------------------------------------------------------------------
func (m *MixerCarry) GetQueryAll() url.Values {
	return m.r.URL.Query()
}
//---------------------------------------------------------------------------
//  возвращает элемент запроса  по ключу url (&{key}={value})
//---------------------------------------------------------------------------
func (m *MixerCarry) GetQuery(key string) string {
	return m.r.URL.Query().Get(key)
}
//---------------------------------------------------------------------------
//  возвращает элемент запроса  по ключу url (&{key}={value})
//---------------------------------------------------------------------------
func (m *MixerCarry) GetFormValue(key string) string {
	return m.r.FormValue(key)
}
func (m *MixerCarry) GetFormValueMultiple(key string) []string {
	m.r.ParseForm()
	return m.r.Form[key]
}
func (m *MixerCarry) GetForm(key string) string {
	return m.r.PostFormValue(key)
}
func (m *MixerCarry) GetCooks() []*http.Cookie {
	return m.r.Cookies()
}
func (m *MixerCarry) GetCook(nameCook string) *http.Cookie {
	for _, c := range m.r.Cookies() {
		if nameCook == c.Name {
			return c
		}
	}
	return nil
}
func (m *MixerCarry) SetCook(c http.Cookie) {
	_newCook := new(http.Cookie)
	_newCook.Name = c.Name
	_newCook.Value = c.Value
	//проверка на domainName из конфига ибо может быть различный доменое имя или IP
	//типа: localhost, 127.0.0.1, 0.0.0.0 - комп один, сетевуха одна, а различие в алиасах есть отсюда разные кукисы
	//установка времени истечения срока действия печеньки

	if time.Now().Sub(c.Expires) > 0 {
		t := time.Now()
		_newCook.Expires = t.Add(time.Duration(86000*30) * time.Minute)
	}
	//установка домена
	if c.Domain == "" {
		domainReal := strings.Split(m.r.Host, ":")[0]
		if "localhost" != domainReal {
			_newCook.Domain = domainReal
		} else {
			_newCook.Domain = "localhost"
		}
	} else {
		_newCook.Domain = c.Domain
	}
	//path
	if c.Path == "" {
		_newCook.Path = "/"
	}
	http.SetCookie(m.w, _newCook)
}
