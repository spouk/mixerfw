package forms

import "errors"

var (
	//---------------------------------------------------------------------------
	//  список дефолтных значений для форм, имеющих соответствующие поля-назания
	//---------------------------------------------------------------------------
	DefaultValues map[string]DefaultForm = map[string]DefaultForm{
		"Name":      DefaultForm{Placeholder: "=введите имя пользователя=", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"Username":  DefaultForm{Placeholder: "=введите имя пользователя=", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"Password":  DefaultForm{Placeholder: "=введите пароль =", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"Email":     DefaultForm{Placeholder: "=введите email =", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"Port":      DefaultForm{Placeholder: "=порт сервера=", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"CatName":   DefaultForm{Placeholder: "=введите название категории=", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"Title":     DefaultForm{Placeholder: "=введите заголовок =", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"MetaKeys":  DefaultForm{Placeholder: "=введите SEO слова =", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"MetaDesc":  DefaultForm{Placeholder: "=введите SEO описание-сниппет =", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"MetaRobot": DefaultForm{Placeholder: "=введите занчения для SEO robot=", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"Message":   DefaultForm{Placeholder: "=введите текст сообщения=", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"Body":      DefaultForm{Placeholder: "=введите тело сообщения =", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"Link":      DefaultForm{Placeholder: "=введите ссылку-ключ =", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"Age":       DefaultForm{Placeholder: "=введите ваш возраст=", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},

		"UserInfoUsername": DefaultForm{Placeholder: "=введите имя пользователя=", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"UserInfoPassword": DefaultForm{Placeholder: "=введите пароль =", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"UserEmail":        DefaultForm{Placeholder: "=введите email =", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"CategoryName":     DefaultForm{Placeholder: "=введите название категории=", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"PostTitle":        DefaultForm{Placeholder: "=введите заголовок =", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"PostBody":         DefaultForm{Placeholder: "=введите тело сообщения =", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"PostSeoMetaKeys":  DefaultForm{Placeholder: "=введите SEO слова =", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"PostSeoMetaDesc":  DefaultForm{Placeholder: "=введите SEO описание-сниппет =", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"PostSeoMetaRobot": DefaultForm{Placeholder: "=введите занчения для SEO robot=", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"TagName":          DefaultForm{Placeholder: "=введите имя метки =", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поле не может быть пустым"},
		"PostCategoryID":   DefaultForm{Placeholder: "", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "значение должно быть выбрано"},
		"PostUserID":       DefaultForm{Placeholder: "", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "значение должно быть выбрано"},
		"Robot":            DefaultForm{Placeholder: "", ErrorClassStyle: "has-error", SuccessClassStyle: "ok", ErrorMsg: "поставьте отметку что вы не робот"},


	}
	//---------------------------------------------------------------------------
	//сообщения об ошибках в формах  
	//---------------------------------------------------------------------------
	ErrorUsername = "Имя пользователя ошибочно"
	ErrorPassword = "Пароль ошибочен"
	ErrorEmail    = "Почтовый адрес ошибочен"

	//---------------------------------------------------------------------------
	//`placeholder` описания для формы  
	//---------------------------------------------------------------------------
	PlaceUsername = "= имя пользователя = "
	PlacePassword = "= пароль ="
	PlaceEmail    = "= почтовый адрес ="

	//---------------------------------------------------------------------------
	//ошибки  
	//---------------------------------------------------------------------------
	ParseErrorInt       = errors.New("[parseform][error] ошибка парсинга `string`->`int64`")
	PTRFormError        = errors.New("[baseform][error] Ошибка, дай мне указатель на структуру для записи")
	PTRFormErrorMethods = errors.New("[baseform][error] Ошибка, отсутствует реализация интерфейса методов для получения данных из формы")
	CSRFErrorValidate   = "CSRF не валидное значение"

	//---------------------------------------------------------------------------
	//название стилей для ошибок в формах полей  
	//---------------------------------------------------------------------------
	ErrorStyleForm   = "has-error"
	SuccessStyleForm = "has-success"

	//---------------------------------------------------------------------------
	//  сообщения для ошибки в формах при валидации формы
	//---------------------------------------------------------------------------
	ErrorMsgFormString   = "- поле не может быть пустым -"
	ErrorMsgFormCheckbox = "- нажмите на чекбокс, если вы не робот -"
	ErrorMsgFormBool     = "- сделайте отметку -"
	ErrorMsgFormSelect   = "- не выбран ни один из элементов -"
	
)
