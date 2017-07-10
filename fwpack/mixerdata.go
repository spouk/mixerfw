package Mixer
//---------------------------------------------------------------------------
//  CONST: константы, текстовые
//---------------------------------------------------------------------------
const (
	//---------------------------------------------------------------------------
	//  CONST:HTTP-MEDIATYPES
	//---------------------------------------------------------------------------
	ApplicationJSON                  = "application/json"
	ApplicationJSONCharsetUTF8       = ApplicationJSON + "; " + CharsetUTF8
	ApplicationJavaScript            = "application/javascript"
	ApplicationJavaScriptCharsetUTF8 = ApplicationJavaScript + "; " + CharsetUTF8
	ApplicationXML                   = "application/xml"
	ApplicationXMLCharsetUTF8        = ApplicationXML + "; " + CharsetUTF8
	ApplicationForm                  = "application/x-www-form-urlencoded"
	ApplicationProtobuf              = "application/protobuf"
	ApplicationMsgpack               = "application/msgpack"
	TextHTML                         = "text/html"
	TextHTMLCharsetUTF8              = TextHTML + "; " + CharsetUTF8
	TextPlain                        = "text/plain"
	TextPlainCharsetUTF8             = TextPlain + "; " + CharsetUTF8
	MultipartForm                    = "multipart/form-data"
	//---------------------------------------------------------------------------
	//  CONST: HTTP-CHARSET
	//---------------------------------------------------------------------------
	CharsetUTF8 = "charset=utf-8"
	//---------------------------------------------------------------------------
	//  CONST:  HTTP-HEADERS
	//---------------------------------------------------------------------------
	AcceptEncoding     = "Accept-Encoding"
	Authorization      = "Authorization"
	ContentDisposition = "Content-Disposition"
	ContentEncoding    = "Content-Encoding"
	ContentLength      = "Content-Length"
	ContentType        = "Content-Type"
	Location           = "Location"
	Upgrade            = "Upgrade"
	Vary               = "Vary"
	WWWAuthenticate    = "WWW-Authenticate"
	XForwardedFor      = "X-Forwarded-For"
	XRealIP            = "X-Real-IP"
	//---------------------------------------------------------------------------
	//  CONST: HTTP-METHODS
	//---------------------------------------------------------------------------
	//RFC 2616
	OPTIONS = "OPTIONS"
	GET     = "GET"
	HEAD    = "HEAD"
	POST    = "POST"
	PUT     = "PUT"
	DELETE  = "DELETE"
	TRACE   = "TRACE"
	CONNECT = "CONNECT"
	//RFC 2518
	PROPFIND  = "PROPFIND"
	PROPPATCH = "PROPPATCH"
	MKCOL     = "MKCOL"
	COPY      = "COPY"
	MOVE      = "MOVE"
	LOCK      = "LOCK"
	UNLOCK    = "UNLOCK"
	//RFC 3253
	REPORT      = "REPORT"
	CHECKOUT    = "CHECKOUT"
	CHECKIN     = "CHECKIN"
	UNCHECKOUT  = "UNCHECKOUT"
	MKWORKSPACE = "MKWORKSPACE"
	UPDATE      = "UPDATE"
	LABEL       = "LABEL"
	MERGE       = "MERGE"
	MKACTIVITY  = "MKACTIVITY"
	
	//---------------------------------------------------------------------------
	//  PREFIX
	//---------------------------------------------------------------------------
	MIXERPREFIX = "[mixer][framework] "
	//---------------------------------------------------------------------------
	//  ERRORS: сообщения об ошибках
	//---------------------------------------------------------------------------
	ERROR_HTTPMETHODNOTACCEPT = MIXERPREFIX + "http method not allowed "
	ERROR_READTEMPLATES = MIXERPREFIX + "%s"
	ERROR_READ_TXTFILE = MIXERPREFIX + "%s"
	//---------------------------------------------------------------------------
	//  MESSAGES: всякие разные информационные сообщения
	//---------------------------------------------------------------------------
	MSG_STARTSERVER  = "starting `%s`\n"
	
)
//---------------------------------------------------------------------------
//  VAR: глобальные переменные
//---------------------------------------------------------------------------
var (
	ACCEPTHTTPMETHOD = []string{
		OPTIONS,
		GET,
		HEAD,
		POST,
		PUT,
		DELETE,
		TRACE,
		CONNECT,
	}
)
