//---------------------------------------------------------------------------
//  Плагин организации работы с формами для Mixerfw 
//  Автор: spouk [spouk@spouk.ru] https://www.spouk.ru
//---------------------------------------------------------------------------
package forms

import (
	"strconv"
	"reflect"
	"strings"
	"fmt"
	"log"
	"database/sql"
	"mixerfw/fwpack"
	"sync"
)

type (
	//структура для дефолтных значений
	DefaultForm struct {
		ErrorMsg          string
		ErrorClassStyle   string
		SuccessClassStyle string
		SuccessMessage    string
		Placeholder       string
	}

	//интерфейс для методов получения данных из формы,
	//для полного абстрагирования от всякого говна, типа фреймворков
	MethodsForms interface {
		GetMultiple(name string) []string
		GetSingle(name string) string
	}

	//интерфейс для юзерских проверок
	UserForm interface {
		Validate(b *MixerForm, stock ...interface{}) bool
	}
	//интерфейс для csrf валидации
	CSRFValidate interface {
		Validate() bool
	}
	//базовая структура для всех форм
	MixerForm struct {
		ParseWithInit bool
		DynamicForm  *DynamicMaps
		CSRFValidate  CSRFValidate
		MethodsForms  MethodsForms
		DefaultForm   map[string]DefaultForm
		pool 	     sync.Pool
	}
	DynamicMaps struct {
		Errors        map[string]string
		Class         map[string]string
		Desc          map[string]string
		Stack         map[string]interface{}
		LastNames     []string
	}
	StockValue struct {
		Name  string
		Value interface{}
	}
	Stocker struct {
		Stock map[string]interface{}
	}
)

//---------------------------------------------------------------------------
//  Стокер структура для внутреннего использования внутри либы
//---------------------------------------------------------------------------
func NewStocker() *Stocker {
	return &Stocker{Stock: make(map[string]interface{})}
}

//---------------------------------------------------------------------------
//  DEFAULT REELEASE METHOD: дефолтная реализация методов получение данных
//---------------------------------------------------------------------------
type MixerMethodsForm struct {
	c *Mixer.MixerCarry
}
func NewMixerMethodsForm(c *Mixer.MixerCarry) MixerMethodsForm {
	return MixerMethodsForm{c:c}
}
func (r MixerMethodsForm) GetMultiple(name string) []string {
	r.c.Request().ParseMultipartForm(0)
	result := r.c.Request().Form[name]
	if result == nil {
		return []string{}
	}
	return result
}
func (r MixerMethodsForm) GetSingle(name string) string {
	result := r.c.Request().FormValue(name)
	return result
}
//---------------------------------------------------------------------------
//  создания инстанса + основной функционал
//---------------------------------------------------------------------------
func NewMixerForm(defaultForms map[string]DefaultForm, methodsForm MethodsForms, ParseWithInit bool) *MixerForm {
	form := new(MixerForm)
	form.ParseWithInit = ParseWithInit
	form.MethodsForms = methodsForm
	form.DefaultForm = defaultForms
	//pool
	form.pool = sync.Pool{}
	//создаю пул с дефолтной функцией
	form.pool.New = func() interface{} {
		return &DynamicMaps{
			Errors:make(map[string]string),
			Class:make(map[string]string),
			Desc:make(map[string]string),
			Stack:make(map[string]interface{}),
			LastNames:[]string{},
		}
	}

	//form.DynamicForm = form.pool.Get().(*DynamicMaps)

	return form
}

//---------------------------------------------------------------------------
//  functions
//---------------------------------------------------------------------------
func (b *MixerForm) ResetForm() {
	if b.DynamicForm != nil {
		b.DynamicForm.Errors = make(map[string]string)
		b.DynamicForm.Class = make(map[string]string)
		b.DynamicForm.Desc = make(map[string]string)
		b.DynamicForm.Stack = make(map[string]interface{})
		//обновляю placeholders
		for _, name := range b.DynamicForm.LastNames {
			if def, exists := b.DefaultForm[name]; exists {
				b.DynamicForm.Desc[name] = def.Placeholder
			}
		}
		b.DynamicForm.LastNames = []string{}
	}
}
func (b *MixerForm) AddForm(name string, form DefaultForm) {
	b.DefaultForm[name] = form
}
func (b *MixerForm) resetErrors() {
	b.DynamicForm.Errors = make(map[string]string)
}

//---------------------------------------------------------------------------
//  INITFORM: инициализация формы дефолтными значениями для этой формы
//---------------------------------------------------------------------------
func (b *MixerForm) InitForm(form UserForm) {

	////создаю новый объект из пула
	b.DynamicForm = b.pool.Get().(*DynamicMaps)

	//проводит первичную инциализацию формы = присваивает placeholder`s
	main := reflect.ValueOf(form)
	numfield := reflect.ValueOf(form).Elem().NumField()
	if main.Kind() != reflect.Ptr {
		log.Fatal(PTRFormError)
	}

	//провожу заполнение
	for x := 0; x < numfield; x++ {
		//получаем имя элемента структуры
		name := reflect.TypeOf(form).Elem().Field(x).Name
		b.DynamicForm.LastNames = append(b.DynamicForm.LastNames, name)
		//получаю placeholder из дефолтного стека
		if def, exists := b.DefaultForm[name]; exists {
			b.DynamicForm.Desc[name] = def.Placeholder
		}
	}
	////провожу парсинг формы //post form
	if b.ParseWithInit {
		b.ParseForm(form)
	}
}
//---------------------------------------------------------------------------
//  UPDATEFORM: обновляет значения формы
//---------------------------------------------------------------------------
func (b *MixerForm) UpdateForm(form UserForm, obj interface{}) {
	if reflect.ValueOf(obj).Kind() != reflect.Ptr || reflect.ValueOf(form).Kind() != reflect.Ptr {
		log.Fatal(PTRFormError)
	}
	//получаю стокер со списком филдов объекта из базы данных как правило
	stocker := b.ParseFields(obj)
	//заполняет форму данными из объекта, используется при обновлении объектов, как пример
	mv := reflect.ValueOf(form).Elem()
	mt := reflect.TypeOf(form).Elem()
	//провожу заполнение
	for x := 0; x < mv.NumField(); x++ {
		//получаем имя элемента структуры
		name := mt.Field(x).Name
		//пробую получить элемент с таким же названием из объекта +
		if v, ok := stocker.Stock[name]; ok {
			b.DynamicForm.Stack[name] = v
		}
	}
}
func (b *MixerForm) ParseFields(obj interface{}) *Stocker {
	//рекурсивно собираю все поля полученном объекте
	stocker := NewStocker()
	for _, x := range b.UpdateFormDeep(obj) {
		stocker.Stock[x.Name] = x.Value
		fmt.Printf("[stock] Name: `%v`     Value : `%v`\n", x.Name, x.Value)
	}
	return stocker
}
func (b *MixerForm) UpdateFormDeep(form interface{}) []StockValue {
	//локальные переменки
	stockFields := make([]StockValue, 0)
	var mv reflect.Value
	var mt reflect.Type

	//form может поступать в 2 видах как указатель так и по значению, отсюда надо ветвить
	switch reflect.ValueOf(form).Kind() {
	case reflect.Ptr:
		mv = reflect.ValueOf(form).Elem()
		mt = reflect.TypeOf(form).Elem()
	default:
		mv = reflect.ValueOf(form)
		mt = reflect.TypeOf(form)
	}
	//дефолтное имя первичноый структуры
	var defaultStructName string
	nameStruct := strings.Split(fmt.Sprintf("%T", mv.Interface()), ".")
	if len(nameStruct) >= 2 {
		defaultStructName = nameStruct[1]
	}
	//рекурсивно получает список всех филдов объекта для последующего отбора
	for x := 0; x < mv.NumField(); x++ {
		v := mv.Field(x)

		switch v.Kind() {
		//case reflect.Struct:
		//
		//	stockFields = append(stockFields, b.UpdateFormDeep(v.Interface())...)
		default:
			s := StockValue{}
			name := mt.Field(x).Name
			s.Name = defaultStructName + name
			//s.Name = name
			s.Value = v.Interface()
			stockFields = append(stockFields, s)
		}
	}
	return stockFields
}
//---------------------------------------------------------------------------
//  PARSEFORM: парсит форму на входящие значения
//---------------------------------------------------------------------------
func (b *MixerForm) ParseForm(obj interface{}) {
	//---------------------------------------------------------------------------
	//  проверка на наличие реализации методово интерфейса для получения данных из формы
	//---------------------------------------------------------------------------
	if b.MethodsForms == nil {
		log.Fatal(PTRFormErrorMethods)
	}
	main := reflect.ValueOf(obj)
	numfield := reflect.ValueOf(obj).Elem().NumField()
	if main.Kind() != reflect.Ptr {
		log.Fatal(PTRFormError)
	}
	//---------------------------------------------------------------------------
	////перебор элементов структуры, получение их имен и получение данных из формы
	//с дальнейшим присваиванием записям структуры
	//---------------------------------------------------------------------------
	for x := 0; x < numfield; x++ {
		//получаем элемент структуры
		f := reflect.Indirect(reflect.ValueOf(obj)).Field(x)
		//получаем имя элемента структуры
		name := reflect.TypeOf(obj).Elem().Field(x).Name

		switch f.Type().Kind() {

		case reflect.Struct:
			value := f.Interface()
			var some interface{}
			some = strings.TrimSpace(b.MethodsForms.GetSingle(name))
			//v := sql.NullString{String:val, Valid:true}
			switch value.(type) {
			//case sql.NullString:
			//	b.Stack[name] = v
			case sql.NullString:
				b.DynamicForm.Stack[name] = some.(sql.NullString).String
			case sql.NullFloat64:
				b.DynamicForm.Stack[name] = some.(sql.NullFloat64).Float64
			case sql.NullInt64:
				b.DynamicForm.Stack[name] = some.(sql.NullInt64).Int64
			case sql.NullBool:
				b.DynamicForm.Stack[name] = some.(sql.NullBool).Bool
			}

		case reflect.Slice, reflect.Array:
			//проводим общие для всех операции
			//получаю данные из формы
			//c.Request().ParseMultipartForm(0)
			//formList2 := c.Request().Form[name]

			formList := b.MethodsForms.GetMultiple(name)
			value := f.Interface()

			switch value.(type) {
			case []int64:
				tmp := []int64{}
				for _, v := range formList {
					if parInt, errPat := strconv.ParseInt(v, 10, 64); errPat == nil {
						tmp = append(tmp, parInt)
					}
				}
				//добавление данных в baseform.stack
				b.DynamicForm.Stack[name] = tmp
				//структура готова, можно менять
				f.Set(reflect.ValueOf(&tmp).Elem())
			case []string:
				tmp := []string{}
				for _, v := range formList {
					tmp = append(tmp, v)
				}
				//добавление данных в baseform.stack
				b.DynamicForm.Stack[name] = tmp
				//меняем
				f.Set(reflect.ValueOf(&tmp).Elem())
			}
		case reflect.Bool:
			val := strings.TrimSpace(b.MethodsForms.GetSingle(name))
			if val != "" {
				f.SetBool(true)
				b.DynamicForm.Stack[name] = true
			} else {
				f.SetBool(false)
				b.DynamicForm.Stack[name] = false
			}

		case reflect.String:
			val := strings.TrimSpace(b.MethodsForms.GetSingle(name))
			f.SetString(val)
			b.DynamicForm.Stack[name] = val

		case reflect.Int, reflect.Int64:
			value := strings.TrimSpace(b.MethodsForms.GetSingle(name))
			if value != "" {
				r, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					log.Printf("%s", ParseErrorInt)
					log.Printf("%s", err)
				} else {
					f.SetInt(r)
					b.DynamicForm.Stack[name] = r
				}
			}
		case reflect.Float64, reflect.Float32:
			value := strings.TrimSpace(b.MethodsForms.GetSingle(name))
			if value != "" {
				r, err := strconv.ParseFloat(value, 64)
				if err != nil {
					log.Printf("%s", ParseErrorInt)
					log.Printf("%s", err)
				} else {
					f.SetFloat(r)
					b.DynamicForm.Stack[name] = r
				}
			}
		}
	}


}
//---------------------------------------------------------------------------
//  ValidateForm: валидация значений формы
//---------------------------------------------------------------------------
func (b *MixerForm) ValidateForm(form UserForm) (status bool) {

	//проверка формы на валидность полей
	main := reflect.ValueOf(form)
	numfield := reflect.ValueOf(form).Elem().NumField()
	if main.Kind() != reflect.Ptr {
		log.Fatal(PTRFormError)
	}
	//обнуляю стек ошибок
	b.resetErrors()
	//проверка CSRF если есть объект реализовавший интерфейс
	if b.CSRFValidate != nil {
		//проверка валидности CSRF значения
		if b.CSRFValidate.Validate() == false {
			b.DynamicForm.Errors["CSRF"] = CSRFErrorValidate
			return false
		}
	}
	//проверка на валидность пользовательской проверкой
	if form.Validate(b) == false {
		return false
	}

	//количество флагов = количество полей в форме
	var total int = numfield
	var countValidate int = 0
	//проверка дефолтных значений и полей
	for x := 0; x < numfield; x++ {
		f := reflect.Indirect(reflect.ValueOf(form)).Field(x)
		name := reflect.TypeOf(form).Elem().Field(x).Name
		ff := reflect.TypeOf(form).Elem().Field(x)

		var def *DefaultForm
		if do, exists := b.DefaultForm[name]; exists {
			def = &do
		}
		switch f.Type().Kind() {
		default:
			fmt.Printf("[reflect][validateform][ALERT] непроверяемый тип Name: `%v` Value: %v\n", name, f.Interface())

		case reflect.Float64, reflect.Float32:
			if ff.Tag != "" {
				total --
			} else {
				result := f.Interface().(float64)
				if result == 0 {
					//error
					if def != nil {
						b.DynamicForm.Class[name] = def.ErrorClassStyle
						b.DynamicForm.Errors[name] = def.ErrorMsg
					}

				} else {
					if def != nil {
						b.DynamicForm.Class[name] = def.SuccessClassStyle
					}
					countValidate ++
				}
			}
		case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int:
			if ff.Tag != "" {
				total --
			} else {
				result := f.Interface().(int64)
				if result == 0 {
					//error
					if def != nil {
						b.DynamicForm.Class[name] = def.ErrorClassStyle
						b.DynamicForm.Errors[name] = def.ErrorMsg
					}

				} else {
					if def != nil {
						b.DynamicForm.Class[name] = def.SuccessClassStyle
					}
					countValidate ++
				}
			}
		case reflect.Slice, reflect.Array:
			//разбор по типу списка
			value := f.Interface()
			switch value.(type) {
			case []int64:
				//проверка на метку необходимости проверки
				//если метка присутствует, проверка не нужна
				if ff.Tag != "" {
					total --
				} else {
					result := value.([]int64)
					if len(result) == 0 {
						//error
						if def != nil {
							b.DynamicForm.Class[name] = def.ErrorClassStyle
							b.DynamicForm.Errors[name] = def.ErrorMsg
						}
					} else {
						if def != nil {
							b.DynamicForm.Class[name] = def.SuccessClassStyle
						}
						countValidate ++
					}
				}
			case []string:

				if ff.Tag != "" {
					total --
				} else {
					result := value.([]string)
					if len(result) == 0 {
						//error
						if def != nil {
							b.DynamicForm.Class[name] = def.ErrorClassStyle
							b.DynamicForm.Errors[name] = def.ErrorMsg
						}
					} else {
						if def != nil {
							b.DynamicForm.Class[name] = def.SuccessClassStyle
						}
						countValidate ++
					}

				}
			}
		case reflect.String:
			tag := ff.Tag
			if tag != "" {
				total --
			} else {
				result := strings.TrimSpace(f.Interface().(string))
				if result == "" {
					//error
					if def != nil {
						b.DynamicForm.Class[name] = def.ErrorClassStyle
						b.DynamicForm.Errors[name] = def.ErrorMsg
					}
					status = false
				} else {
					if def != nil {
						b.DynamicForm.Class[name] = def.SuccessClassStyle
					}

					countValidate ++
				}
			}

		case reflect.Bool:
			tag := ff.Tag
			if tag != "" {
				total --
			} else {
				result := f.Interface().(bool)
				if result == false {
					//error
					if def != nil {
						b.DynamicForm.Class[name] = def.ErrorClassStyle
						b.DynamicForm.Errors[name] = def.ErrorMsg
					}

				} else {
					if def != nil {
						b.DynamicForm.Class[name] = def.SuccessClassStyle
					}
					countValidate ++
				}
			}
		}

	}

	//подведение итогов по валидности всей формы
	if total == countValidate {
		fmt.Printf("[validateform] Total: %v   Numfield: %v   CountValidate: %v , Result: VALIDATE\n", total, numfield, countValidate)
		status = true
	} else {
		status = false
		fmt.Printf("[validateform] Total: %v   Numfield: %v   CountValidate: %v , Result: NOT VALIDATE\n", total, numfield, countValidate)
	}
	//возвращаем динамический мапы в пул
	b.pool.Put(b.DynamicForm)

	return
}
func (b *MixerForm) ConvertSliceINT64(name string) []int64 {
	v := b.DynamicForm.Stack[name]
	result := []int64{}
	if v != nil {
		result = v.([]int64)
	}
	return result
}
func (b *MixerForm) ConvertString(name string) string {
	v := b.DynamicForm.Stack[name]
	var result string
	if v != nil {
		result = v.(string)
	}
	return result
}
func (b *MixerForm) ConvertInt(name string) int64 {
	v := b.DynamicForm.Stack[name]

	var result int64
	if v != nil {
		result = int64(v.(int))
	}
	return result
}
func (b *MixerForm) ConvertBool(name string) bool {
	v := b.DynamicForm.Stack[name]
	var result bool
	if v != nil {
		result = v.(bool)
	}
	return result
}
