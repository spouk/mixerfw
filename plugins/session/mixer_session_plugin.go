//---------------------------------------------------------------------------
//  Плагин организации сессии для Mixerfw в виде миддла
//  Автор: spouk [spouk@spouk.ru] https://www.spouk.ru
//---------------------------------------------------------------------------
package session

import (
	"sync"
	"time"
	"fmt"
	"crypto/md5"
	"crypto/sha1"
	"math/rand"
	"mixerfw/fwpack"
	"mixerfw/plugins/database"
	"os"
	"net/http"
	"crypto/tls"
	gml "gopkg.in/gomail.v2"
	d "github.com/fiam/gounidecode/unidecode"
	"errors"
	"math"
	"strings"
	"reflect"
	"strconv"
)

const (
	//---------------------------------------------------------------------------
	//  LETTERS STRING
	//---------------------------------------------------------------------------
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	//---------------------------------------------------------------------------
	//  ROLE
	//---------------------------------------------------------------------------
	DEFAULT_USER_ROLE = "anonymous"
	//---------------------------------------------------------------------------
	//  SESSION DEFAULT TIME
	//---------------------------------------------------------------------------
	SESSION_DEFAULT_TIME = 5
	//---------------------------------------------------------------------------
	//  PREFIX SESSION
	//---------------------------------------------------------------------------
	PREFIX_SESSION     = "[mixer][session] "
	ERROR_MSG_ADDUSER  = PREFIX_SESSION + "%s"
	ERROR_MSG_SAVECOOK = PREFIX_SESSION + "%s"
	//---------------------------------------------------------------------------
	//  DATA: дефолтные секции
	//---------------------------------------------------------------------------
	SECTION_ADMIN  = "admin"
	SECTION_PUBLIC = "public"
	SECTION_USER   = "user"
	SECTION_FLASH  = "flash"
	SECTION_STACK  = "stack"
	SECTION_SEO    = "seo"
	SECTION_FORM   = "form"
	//---------------------------------------------------------------------------
	//  FLASH:
	//---------------------------------------------------------------------------
	SALT_FLASH_HASH = "#$%dfgdfgWERfvdfgdfgFg"
	//---------------------------------------------------------------------------
	//  CSRF
	//---------------------------------------------------------------------------
	CSRF_LENGTH_KEY  = 7
	CSRF_SALT        = "Cvdfg345DFg234@#$dfgxcvbq"
	CSRF_ACTIVE_TIME = 1
	//---------------------------------------------------------------------------
	//  COOKIE
	//---------------------------------------------------------------------------
	COOKIE_SALT = "dfg$%^FGhfgh456FGHfgh"
)

var (
	DATA_SECTION = []string{SECTION_ADMIN, SECTION_PUBLIC, SECTION_USER, SECTION_FLASH,
				SECTION_SEO, SECTION_STACK, SECTION_FORM}
)

type (
	//---------------------------------------------------------------------------
	//  TABLES FOR WORK SESSION: таблицы для сессии
	//---------------------------------------------------------------------------
	MixerCookTable struct {
		ID           int64 `gorm:"primary_key"`
		Created      int64
		Cookie       string `sql:"unique"`
		Dump         string
		Status       bool //если true то показывает что пользователь реальный а не анонимус
		Logged       bool
		Lastconnect  int64 //time.UnixNano
		Remoteaddr   string `sql:"type:varchar(1000)"`
		UserAgent    string `sql:"type:varchar(5000); "`
		CountConnect int64
		Refer        string
	}
	//---------------------------------------------------------------------------
	//  SESSION:
	//---------------------------------------------------------------------------
	MixerSession struct {
		Mail     *MixerMail
		Paginate *MixerPaginate
		Convert  *MixerConvert
		CSRF     *MixerCSRF
		DBS      *database.MixerDBS
		info     *MixerSessionInfo
		pool     sync.Pool
		Logger   Mixer.MixerLogger
		Flash   *MixerFlash
		Transliter *MixerTransliter
	}
	MixerSessionInfo struct {
		//database
		DBS *database.MixerDBS
		//mail info
		MailTo       string
		MailFrom     string
		MailHost     string
		MailPort     int
		MailUsername string
		MailPassword string
		//csrf
		CSRFTimeActive int
		CSRFSalt       string
		//cookie
		CookieName    string
		CookieDomain  string
		CookieExpired int64
		CookieSalt    string
		//session time
		SessionTime int64 //minutes
		CheckSessionExpire int //second(debug) а так минуты
		//role роль, для иерархии пользователей + разграничение по группам
		RoleDefaultUser string
	}
	MixerSessionCarry struct {
		Data    *MixerDATA
		Flash   *MixerFlash
		Session *MixerSession
		Cook    *MixerCookTable
	}
	//---------------------------------------------------------------------------
	//  MAIL: поддержка почты в рамках сессии для обработчиков запросов
	//---------------------------------------------------------------------------
	MixerMail struct {
		MailMessage MixerMailMessage

	}
	MixerMailMessage struct {
		To         string
		From       string
		Message    string
		Subject    string
		FileAttach string `fullpath to file`
		Host       string
		Port       int
		Username   string
		Password   string
	}

	//---------------------------------------------------------------------------
	//  FLASH: поддержка контейнера коротких сообщений в рамках сессии
	//---------------------------------------------------------------------------
	MixerFlash struct {
		sync.RWMutex
		Key   string
		Stock map[string]*MixerFlashMessage
	}
	MixerFlashMessage struct {
		Status  string
		Message interface{}
	}
	//---------------------------------------------------------------------------
	//  TRANSLITE: транслитерация
	//---------------------------------------------------------------------------
	MixerTransliter struct {
		Validate []int
		Replacer []int
		InValid  []int
	}
	//---------------------------------------------------------------------------
	//  PAGINATE: пагинация результатов базы данных для представления
	//---------------------------------------------------------------------------
	MixerPaginate struct {
	}
	//---------------------------------------------------------------------------
	//  CONVERT: поддержка конвертации разных величин
	//---------------------------------------------------------------------------
	MixerConvert struct {
		logger  Mixer.MixerLogger
		value   interface{}
		result  interface{}
		stockFu map[string]func()		
	}
	//---------------------------------------------------------------------------
	//  CSRF: добавление поддержки
	//---------------------------------------------------------------------------
	MixerCSRF struct {
		TimeActive time.Duration
		TimeStart  time.Time
		Salt       string
		Key        string
		ReadyKey   string
		Csrf_form  func() (*string, error)
		Csrf_head  func() (*string, error)
	}
	//---------------------------------------------------------------------------
	//  DATA: динамический контейнер в рамках контекста
	//---------------------------------------------------------------------------
	MixerDATA map[string]interface{}
)

//---------------------------------------------------------------------------
//  TRANSLITER
//---------------------------------------------------------------------------
func NewMixerTransliter() *MixerTransliter {
	//	hex 65-122 A-z (допустимые )
	//	hex 48-57 0-9 ( допустимые )
	// 	hex 20 (зменяемые)
	//	hex 123-126, 91-96, 58-64, 33-47 punctuations ( запретные )
	n := new(MixerTransliter)
	n.Validate = n.convert(65, 122)
	n.InValid = n.convert(123, 126)

	n.InValid = append(n.InValid, n.convert(91, 96)...)
	n.InValid = append(n.InValid, n.convert(58, 64)...)
	n.InValid = append(n.InValid, n.convert(33, 47)...)
	n.InValid = append(n.InValid, n.convert(33, 47)...)
	n.Replacer = n.convert(32, 32)
	//n.Replacer = append(n.Replacer, 20)
	return n
}
func (s *MixerTransliter) convert(start, end int) []int {
	stack := []int{}
	for ; start <= end; start++ {
		stack = append(stack, start)
	}
	return stack
}
func (s *MixerTransliter) correct(str string) string {

	var result []string
	for _, x := range strings.Split(strings.TrimSpace(str), " ") {
		if x != "" {
			result = append(result, x)
		}
	}
	return strings.Join(result, " ")
}
func (s *MixerTransliter) preCorrect(str string) string {
	str = s.correct(str)
	var tmp []string
	for _, sym := range str {
		switch {
		case s.InSlice(s.InValid, int(sym)): continue
		case s.InSlice(s.Validate, int(sym)): tmp = append(tmp, string(sym))
		case s.InSlice(s.Replacer, int(sym)): tmp = append(tmp, " ")
		default: tmp = append(tmp, string(sym))
		}
	}
	return s.correct(strings.Join(tmp, ""))
}

func (s *MixerTransliter) TransliterCyr(str string) string {
	str = s.preCorrect(str)
	var result []string
	for _, sym := range d.Unidecode(str) {
		switch {
		case s.InSlice(s.InValid, int(sym)):
			continue
		case s.InSlice(s.Validate, int(sym)):
			result = append(result, string(sym))
		case s.InSlice(s.Replacer, int(sym)):
			result = append(result, "-")
		default:
			result = append(result, string(sym))
		}
	}
	return strings.Join(result, "")
}
func (s *MixerTransliter) InSlice(str []int, target int) bool {
	for x := 0; x < len(str); x++ {
		if str[x] == target {
			return true
		}
	}
	return false
}
func (s *MixerTransliter) ShowAscii() {
	var i int
	for i = 0; i < 255; i++ {
		fmt.Printf("Dec: %3d Sym: %3c Hex: %3x\n", i, i, i)
	}
}
//---------------------------------------------------------------------------
//  MAIL
//---------------------------------------------------------------------------
func NewMixerMail() *MixerMail {
	return &MixerMail{MailMessage:MixerMailMessage{}}
}
func (mail MixerMailMessage) SendMail(message *MixerMailMessage) (error) {
	d := gml.NewPlainDialer(message.Host, message.Port, message.Username, message.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	m := gml.NewMessage()
	m.SetHeader("From", message.From)
	m.SetHeader("To", message.To)
	//	m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", message.Subject)
	m.SetBody("text/html", message.Message)
	if message.FileAttach != "" {
		m.Attach(message.FileAttach)
	}
	if err := d.DialAndSend(m); err != nil {
		fmt.Printf("[sendemail] ошибка отправки сообщения %v\n", err)
		return errors.New(fmt.Sprintf("[sendemail] ошибка отправки сообщения %v\n", err))
	}
	return nil
}
//---------------------------------------------------------------------------
//  DATA:
//---------------------------------------------------------------------------
func NewMixerData() *MixerDATA {
	d := make(MixerDATA)
	d.setDefaultSection()
	return &d
}
func (s MixerDATA) setDefaultSection() {
	for _, x := range DATA_SECTION {
		s[x] = make(map[string]interface{})
	}
}
func (s MixerDATA) Set(section, key string, value interface{}) {
	s[section].(map[string]interface{})[key] = value
}
func (s MixerDATA) Get(section, key string) (interface{}) {
	return s[section].(map[string]interface{})[key]
}

//---------------------------------------------------------------------------
//  FLASH:
//---------------------------------------------------------------------------
func NewMixerFlash() *MixerFlash {
	n := &MixerFlash{
		Stock: make(map[string]*MixerFlashMessage),
	}
	n.Key = n.generateKey()
	return n
}
func (f *MixerFlash) generateKey() string {
	t := time.Now()
	return fmt.Sprintf("%x", md5.Sum([]byte(t.String()+SALT_FLASH_HASH)))
}
func (f *MixerFlash) Set(status, section string, message interface{}) {
	nm := &MixerFlashMessage{Status: status, Message: message}
	f.Lock()
	f.Stock[section] = nm
	f.Unlock()
}
func (f *MixerFlash) Get(section string) (*MixerFlashMessage) {
	f.Lock()
	defer f.Unlock()
	if result, exists := f.Stock[section]; exists {
		delete(f.Stock, section)
		return result
	}
	return nil
}
func (f *MixerFlash) HaveMsg() bool {
	if len(f.Stock) > 0 {
		return true
	}
	return false
}

//---------------------------------------------------------------------------
// CSRF:
//---------------------------------------------------------------------------
func NewMixerCSRF(minutesActive int, salt string) (*MixerCSRF) {
	n := &MixerCSRF{
		TimeActive: time.Duration(minutesActive) * time.Minute,
		TimeStart:  time.Now(),
		Salt:       salt,
	}
	n.Key = n.randomGenerate(CSRF_LENGTH_KEY)
	n.Csrf_form = n.wrapper(true, false)
	n.Csrf_head = n.wrapper(false, true)
	return n
}
func (c *MixerCSRF) wrapper(form, head bool) (func() (*string, error)) {
	return func() (*string, error) {
		_tmptime := c.TimeStart.Add(c.TimeActive)
		if _tmptime.Before(time.Now()) {
			c.Key = c.randomGenerate(CSRF_LENGTH_KEY)
		}
		r := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%v%v", c.Key, c.Salt))))
		c.ReadyKey = r
		var result string
		if form {
			result = fmt.Sprintf(`<input type="hidden" name="csrf_token" value="%s"> `, r)
		} else if head {
			result = fmt.Sprintf(`<meta id="csrf_token_ajax" content="%s" name="csrf_token_ajax" />`, r)
		}
		return &result, nil
	}
}
func (c MixerCSRF) VerifyToken(s *Mixer.MixerCarry) bool {
	token := s.GetFormValue("csrf_token")
	if token == c.ReadyKey {
		return true
	}
	return false
}
func (c MixerCSRF) VerifyTokenString(token string) bool {
	if token == c.ReadyKey {
		return true
	}
	return false
}
func (c MixerCSRF) randomGenerate(count int) string {
	randInt := func(min, max int) int {
		return min + rand.Intn(max-min)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	bytes := make([]byte, count)
	for i := 0; i < count; i++ {
		bytes[i] = byte(randInt(30, 90))
	}
	h := sha1.New()
	h.Write(bytes)
	return fmt.Sprintf("%x", h.Sum(nil))
}

//---------------------------------------------------------------------------
//  SESSION:
//---------------------------------------------------------------------------
func NewMixerSession(info MixerSessionInfo) *MixerSession {
	s := new(MixerSession)
	s.info = &info
	//проверяю есть ли коннектор к базе данных
	if info.DBS != nil {
		s.DBS = info.DBS
		//создаю таблицу с куками, если ее нет в базе данных
		s.DBS.CreateTablesInDatabase([]interface{}{MixerCookTable{}})
	}
	//создаю CSRF
	var csrfsalt string
	if info.CSRFSalt == "" {
		csrfsalt = info.CSRFSalt
	} else {
		csrfsalt = CSRF_SALT
	}
	if info.CSRFTimeActive > 0 {
		s.CSRF = NewMixerCSRF(info.CSRFTimeActive, csrfsalt)
	} else {
		s.CSRF = NewMixerCSRF(CSRF_ACTIVE_TIME, csrfsalt)
	}
	//создаю логгер
	s.Logger = Mixer.NewMixerLogger(PREFIX_SESSION, os.Stdout)

	//создаю статичный флешер для сессии и передачи его в SessionCarry
	s.Flash = NewMixerFlash()

	//создаю пул
	s.pool = sync.Pool{}
	s.pool.New = func() interface{} {
		return &MixerSessionCarry{
			Data:    NewMixerData(),
			Flash:	s.Flash,
			Session: s,
		}
	}
	//mail
	s.Mail = NewMixerMail()

	//запускаю горутину для отслеживания и удаления истекшие сессии(кукисы)
	go s.foundExpiredSessionCooks(s.info.CheckSessionExpire)

	//создаю конвертер
	s.Convert = s.NewMixerConverter(s.Logger)

	//создаю флешер
	s.Flash = NewMixerFlash()

	//создаю транслитер
	s.Transliter = NewMixerTransliter()

	return s
}
func (s *MixerSession) ShowInfo() *MixerSessionInfo{
	return s.info
}
//---------------------------------------------------------------------------
//  SESSION: работа с кукисами и сессиями
//---------------------------------------------------------------------------
//проверка сессии на наличие как активной
func (s *MixerSession) foundExpiredSessionCooks(timeCheck int) {

	s.Logger.Printf("`foundExpiredSessionCooks` running...")
	var (
		cookies      []MixerCookTable
		sessiontimer int64
	)
	if s.info.SessionTime == 0 {
		//если отсутствует время сессии, ставлю дефолтное время
		sessiontimer = SESSION_DEFAULT_TIME
	} else {
		sessiontimer = s.info.SessionTime
	}
	//бесконечный цикл т.к. эта функция пойдет в горутине
	for {
		//s.Logger.Printf("%s\n", "expiredSession starting....")
		if err := s.DBS.DB.Where("lastconnect + ? < ?", sessiontimer * 60, time.Now().Unix()).Find(&cookies).Error; err != nil {
			s.Logger.Printf(ERROR_MSG_SAVECOOK, err.Error())
		} else {
			//удаляю найденные истекшие куки
			if len(cookies) > 0 {
				var realdeletecook int = 0
				for _, x := range cookies {
					//если НЕ активный пользователь ( НЕ имеющий регистрацию,мыло,пароль, имя и прочее)
					//реального пользователя отличает выставленный флаг .Status
					if !x.Status {
						if err_del := s.DBS.DB.Delete(x).Error; err_del != nil {
							s.Logger.Printf(ERROR_MSG_SAVECOOK, err_del.Error())
						}
						realdeletecook++
					} else {
						//если пользователь активный то выставляем false в logged
						x.Logged = false
						s.DBS.DB.Save(x)
					}
				}
				s.Logger.Printf("deleting expired session `%3d`\n", realdeletecook)
			}
		}
		//время спать
		time.Sleep(time.Duration(timeCheck) * time.Minute)
	}
}

//проверка кукиса на наличие в базе данных
func (s *MixerSession) checkCookandMakeNewCook(m *Mixer.MixerCarry)  *MixerCookTable{
	//получаю кукис
	cook := m.GetCook(s.info.CookieName)
	if cook != nil && cook.Value != "" {
		//кукус найден,проверяем по базе
		c := &MixerCookTable{}
		if err := s.DBS.DB.Find(c, MixerCookTable{Cookie: cook.Value}).Error; err == nil {
			//кукис найден в базе данных
			//обновляю время `lastconnect`
			c.Lastconnect = time.Now().Unix()
			if err := s.DBS.DB.Save(c).Error; err != nil {
				s.Logger.Printf(ERROR_MSG_SAVECOOK, err.Error())
			}
			//s.Cook = c
			return c
		}
	}
	//кукиса не найдено, создаем новый кукис
	newuser := s.adduser(m)
	if newuser != nil {
		//успешно создан
		newcook := http.Cookie{}
		newcook.Name = s.info.CookieName
		newcook.Value = newuser.Cookie
		//устанавливаю новый кукис новому юзеру
		m.SetCook(newcook)
		return newuser
	}
	return nil
}
func (s *MixerSession) SetNewCookie(newcookie string, m *Mixer.MixerCarry) {
	newcook := http.Cookie{}
	newcook.Name = s.info.CookieName
	newcook.Value = newcookie
	m.SetCook(newcook)
}
func (s *MixerSession) Cookgeneratenew() string {
	//генерация нового хэша
	t := time.Now()
	result := fmt.Sprintf("%x", md5.Sum([]byte(t.String() + s.info.CookieSalt)))
	return result
}
func (s *MixerSession) AddNewCook(m *Mixer.MixerCarry) *MixerCookTable {
	return s.adduser(m)
}
//добавление нового кукиса
func (s *MixerSession) adduser(m *Mixer.MixerCarry) *MixerCookTable {
	newuser := &MixerCookTable{}
	s.Logger.Printf("[session] Info: %v\n", s.info)
	newuser.Cookie = s.Cookgeneratenew()
	//newuser.Email = fmt.Sprintf("anonymous@%s.ru", newuser.Cookie)
	newuser.Created = time.Now().Unix()
	newuser.Lastconnect = time.Now().Unix()
	newuser.Status = false
	newuser.CountConnect = 1
	newuser.Remoteaddr = m.Request().RemoteAddr
	newuser.UserAgent = m.Request().UserAgent()
	newuser.Refer = m.Request().Header.Get("Referer")
/*	//if s.info.RoleDefaultUser == "" {
	//	newuser.Role = DEFAULT_USER_ROLE
	//} else {
	//	newuser.Role = s.info.RoleDefaultUser
	//}*/
	s.Logger.Printf("[SESSION] adduser : %v\n", newuser)
	if err := s.DBS.DB.Create(newuser).Error; err != nil {
		//error create new user
		s.Logger.Printf(ERROR_MSG_ADDUSER, err.Error())
		return nil
	}
	return newuser
}
//---------------------------------------------------------------------------
//  SESSION: middleware
//---------------------------------------------------------------------------
func (s *MixerSession) MixerSessionMiddleware(handler Mixer.MixerHandler) Mixer.MixerHandler {
	return Mixer.MixerHandler(func(c *Mixer.MixerCarry) error {
		//проверяю кукис входящий
		cook := s.checkCookandMakeNewCook(c)
		//создаю новый объект сессионной несущей для каждого запроса свой контейнер
		ss := s.pool.Get().(*MixerSessionCarry)
		ss.Cook = cook
		//добавляю сессионный контейнер к текущему контексту
		ss.Data.Set("stack", "ss", ss)
		c.Set("session", ss)
		handler(c)
		return nil
	})
}
//---------------------------------------------------------------------------
//  SESSION: PAGINATE
//---------------------------------------------------------------------------
func (s *MixerSession) PaginateHTML(current_page int, count_on_page int, count_links int, path string, total_len int) string {
	// start,middle,end; middle = stacklinks, start,end = управляющие кнопки
	if current_page == 0 {
		current_page = 1
	}
	current_page = int(current_page)
	var tmp string = `<ul class="pagination"> %s </ul>`
	var START string = `<li><a class="%s" href="%s/page/%d" %s> < </a></li>`
	var END string = `<li><a class="%s" href="%s/page/%d" %s> > </a></li>`
	var LINK string = `	<li><a class="%s" href="%s/page/%d">%d</a></li>`
	var LINK_ACTIVE string = `<li><a class="%s" href="%s/page/%d">%d</a></li>`
	var out string
	var stacklinks string
	var start, end string
	var plinks []string
	//общее количество страниц исходя из количества элементов на странице
	totalpages := int(math.Ceil(float64(total_len) / float64(count_on_page)))

	for i := 1; i <= totalpages; i++ {
		if i == current_page {
			plinks = append(plinks, fmt.Sprintf(LINK_ACTIVE, "active", path, i, i))
		} else {
			plinks = append(plinks, fmt.Sprintf(LINK, "", path, i, i))
		}
	}
	// если общее количество страниц `num_pages` <= количеству страниц показывемых в пагинации
	//start,end страницы выставляются в disabled
	if totalpages <= count_links {
		if current_page == 1 {
			start = fmt.Sprintf(START, "disabled", path, current_page, "disabled")
		} else {
			start = fmt.Sprintf(START, "", path, current_page - 1, "")
		}

		if current_page < totalpages {
			end = fmt.Sprintf(END, "", path, current_page + 1, "")
		} else {
			end = fmt.Sprintf(END, "disabled", path, current_page, "disabled")
		}
	}
	if totalpages > count_links {
		if totalpages - current_page >= count_links {
			plinks = plinks[current_page - 1: current_page + count_links - 1]
			//fmt.Printf("-- total > countlinks\n", plinks)
		} else {
			//fmt.Printf("-- total < countlinks\n", plinks)
			plinks = plinks[totalpages - count_links:]
		}
		if current_page == 1 {
			start = fmt.Sprintf(START, "disabled", path, current_page, "disabled")
			end = fmt.Sprintf(END, "", path, current_page + 1, "")
		}
		if current_page > 1 {
			start = fmt.Sprintf(START, "", path, current_page - 1, "")
			if totalpages == current_page {
				end = fmt.Sprintf(END, "disabled", path, current_page + 1, "disabled")
			} else {
				end = fmt.Sprintf(END, "", path, current_page + 1, "")
			}
		}
	}
	//compose
	stacklinks += start
	stacklinks += strings.Join(plinks, "")
	stacklinks += end
	out = fmt.Sprintf(tmp, stacklinks)
	return out
}
//---------------------------------------------------------------------------
//  SESSION: CONVERT
//---------------------------------------------------------------------------
const (
	defConverter = "[mixerconvert] `%s`\n"
	prefixLogConverter  = "[mixerconvert][logger]"
	ErrorValueNotValidConvert = "не подходящее значение для конвертации"

)
var (
	acceptTypes []interface{} = []interface{}{
		"", 0, int64(0),
	}
)
func (m *MixerSession) NewMixerConverter(log Mixer.MixerLogger) *MixerConvert {
	f := &MixerConvert{
		stockFu:make(map[string]func()),
		logger: log,
	}
	f.stockFu["string"] = f.stringToInt
	f.stockFu["string"] = f.stringToInt64
	return f
}
func (m *MixerConvert) StrToInt() (*MixerConvert) {
	if f, exists := m.stockFu["string"]; exists {
		f()
	}
	return m
}
func (m *MixerConvert) StrToInt64() (*MixerConvert) {
	if f, exists := m.stockFu["string"]; exists {
		f()
	}
	return m
}
//---------------------------------------------------------------------------
//  String to Int64
//---------------------------------------------------------------------------
func (m *MixerConvert) stringToInt64() {
	m.stringToInt()
	if m.result != nil {
		m.result = int64(m.result.(int))
	} else {
		m.result = nil
	}
}
//---------------------------------------------------------------------------
//  String to Int
//---------------------------------------------------------------------------
func (m *MixerConvert) stringToInt() {
	if r, err := strconv.Atoi(m.value.(string)); err != nil {
		m.logger.Printf(defConverter, err.Error())
		m.result = nil
	} else {
		m.result = r
	}
}
//---------------------------------------------------------------------------
//  возвращает результат конвертации
//---------------------------------------------------------------------------
func (m *MixerConvert) Result() interface{} {
	return m.result
}
//---------------------------------------------------------------------------
//  инциализация вводным значением
//---------------------------------------------------------------------------
func (m *MixerConvert) Value(value interface{}) (*MixerConvert) {
	if m.checkValue(value) {
		m.value = value
		return m
	}
	return nil
}
//---------------------------------------------------------------------------
//  проверка типа поступившего значения на возможность конвертации
//---------------------------------------------------------------------------
func (m *MixerConvert) checkValue(value interface{}) bool {
	tValue := reflect.TypeOf(value)
	for _, x := range acceptTypes {
		if tValue == reflect.TypeOf(x) {
			return true
		}
	}
	m.logger.Printf(defConverter, ErrorValueNotValidConvert)
	return false
}

func (m *MixerConvert) FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 2, 64)

}
func (m *MixerConvert) Int64ToString(input_num int64) string {
	return strconv.FormatInt(input_num, 10)

}
func (m *MixerConvert) DirectStringtoInt64(v string) int64 {
	if res, err := strconv.Atoi(v); err != nil {
		m.logger.Printf(defConverter, err.Error())
		return 0
	} else {
		return int64(res)
	}
}
func (m *MixerConvert) DirectStringtoFloat64(v string) float64 {
	if res, err := strconv.ParseFloat(v, 10); err != nil {
		m.logger.Printf(defConverter, err.Error())
		return 0
	} else {
		return res
	}
}
//---------------------------------------------------------------------------
//  time convert
//---------------------------------------------------------------------------
func (m *MixerConvert) ConvertHTMLDatetoUnix(date string) (int64, error) {
	result, err := time.Parse("2006-01-02", date)
	if err == nil {
		return result.Unix(), err
	} else {
		return 0, err
	}
}
func (m *MixerConvert) ConvertUnixTimeToString(unixtime int64) string {
	return time.Unix(unixtime, 0).String()
}
//convert timeUnix->HTML5Datatime_local(string)
func (m *MixerConvert) TimeUnixToDataLocal(unixtime int64) string {
	tmp_result := time.Unix(unixtime,0).Format(time.RFC3339)
	g := strings.Join(strings.SplitAfterN(tmp_result,":",3)[:2],"")
	return g[:len(g)-1]
}
//convert HTML5Datatime_local(string)->TimeUnix
func (m *MixerConvert) DataLocalToTimeUnix(datatimeLocal string) int64 {
	r,_ := time.Parse(time.RFC3339, datatimeLocal+":00Z")
	return r.Unix()
}
//convert HTML5Data->UnixTime
func (m *MixerConvert) HTML5DataToUnix(s string) int64 {
	l := "2006-01-02"
	r , _ := time.Parse(l, s)
	return r.Unix()
}
//UnixTime->HTML5Data
func  (m *MixerConvert) UnixtimetoHTML5Date(unixtime int64) string {
	return time.Unix(unixtime, 0).Format("2006-01-02")
}
//рандомный генератор строк переменной длины
func (m *MixerConvert) RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
