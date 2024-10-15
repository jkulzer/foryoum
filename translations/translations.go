package translations

import (
	"net/http"
)

type Translations struct {
	MainPage                  string
	Posts                     string
	Search                    string
	SearchAction              string
	Register                  string
	Login                     string
	Logout                    string
	Sessions                  string
	Username                  string
	Password                  string
	CommentingMessage         string
	CommentButton             string
	IfYouWantToLeaveAComment  string
	RegisterLowerCase         string
	Or                        string
	LoginLowerCase            string
	Attachments               string
	CreateAPost               string
	Title                     string
	CreatePost                string
	FilesizeLimit             string
	Preview                   string
	User                      string
	Says                      string
	ByUser                    string
	Next                      string
	Previous                  string
	FailedToRegister          string
	SearchForPosts            string
	CreatedAt                 string
	Token                     string
	Delete                    string
	InvalidPostRange          string
	InvalidPostID             string
	InvalidSearchRange        string
	InvalidAttachmentLocation string
	FailedToParseFormData     string
	FailedToAddAttachments    string
	PostedSucessfully         string
	FailedToLoadPosts         string
	UsernameAlreadyTaken      string
	Content                   string
	CommentPlaceholder        string
	Comments                  string
}

var translations = map[string]Translations{
	"EN": {
		MainPage:                  "Main Page",
		Posts:                     "Posts",
		Search:                    "Search",
		SearchAction:              "Search!",
		Register:                  "Register",
		Login:                     "Login",
		Logout:                    "Logout",
		Sessions:                  "Sessions",
		Username:                  "Username",
		Password:                  "Password",
		CommentingMessage:         "Contribute to the discussion!",
		CommentButton:             "Comment!",
		IfYouWantToLeaveAComment:  "If you want to leave a comment, ",
		RegisterLowerCase:         "register",
		Or:                        "or ",
		LoginLowerCase:            "log in",
		Attachments:               "Attachments",
		CreateAPost:               "Create a post",
		Title:                     "Title",
		CreatePost:                "Create post",
		FilesizeLimit:             "(filesize limit of 10MB)",
		Preview:                   "Preview",
		User:                      "User",
		Says:                      "says",
		ByUser:                    "by user",
		Next:                      "Next",
		Previous:                  "Previous",
		FailedToRegister:          "Failed to register",
		SearchForPosts:            "Search for posts...",
		CreatedAt:                 "Created at ",
		Token:                     "Token",
		Delete:                    "Delete",
		InvalidPostRange:          "Invalid post range",
		InvalidPostID:             "Invalid post ID",
		InvalidSearchRange:        "Invalid search range",
		InvalidAttachmentLocation: "Invalid attachment location",
		FailedToParseFormData:     "Failed to parse form data",
		FailedToAddAttachments:    "Failed to add attachments",
		PostedSucessfully:         "Posted sucessfully!",
		FailedToLoadPosts:         "Failed to load posts",
		UsernameAlreadyTaken:      "Username already taken",
		Content:                   "Put your post content here! \nThis box accepts Markdown-formatted text!",
		CommentPlaceholder:        "Write your comment here!",
		Comments:                  "Comments",
	},
	"DE": {
		MainPage:                  "Startseite",
		Posts:                     "Beiträge",
		Search:                    "Suche",
		SearchAction:              "Suchen",
		Register:                  "Registrieren",
		Login:                     "Anmeldung",
		Logout:                    "Abmelden",
		Sessions:                  "Sitzungen",
		Username:                  "Benutzername",
		Password:                  "Passwort",
		CommentingMessage:         "Beteilige dich an der Diskussion!",
		CommentButton:             "Kommentieren",
		IfYouWantToLeaveAComment:  "Wenn du kommentieren willst, ",
		RegisterLowerCase:         "registrier dich",
		Or:                        " oder ",
		LoginLowerCase:            "melde dich an",
		Attachments:               "Anhänge",
		CreateAPost:               "Erstelle einen Post",
		Title:                     "Titel",
		CreatePost:                "Beitrag erstellen",
		FilesizeLimit:             "(Maximale Dateigröße: 10MB)",
		Preview:                   "Vorschau",
		User:                      "Benutzer",
		Says:                      "sagt",
		ByUser:                    "von Benutzer",
		Next:                      "Weiter",
		Previous:                  "Zurück",
		FailedToRegister:          "Registrierung fehlgeschlagen",
		SearchForPosts:            "Nach Beiträgen suchen...",
		CreatedAt:                 "Erstellt am Zeitpunkt ",
		Token:                     "Token",
		Delete:                    "Löschen",
		InvalidPostRange:          "Ungültiger Index",
		InvalidPostID:             "Ungültige Post-ID",
		InvalidSearchRange:        "Ungültiger Index",
		InvalidAttachmentLocation: "Ungültiger Anlagenpfad",
		FailedToParseFormData:     "Auslese der Formulardaten fehlgeschlagen",
		FailedToAddAttachments:    "Anhänge konnten nicht hinzugefügt werden",
		PostedSucessfully:         "Erstellung des Beitrags erfolgreich!",
		FailedToLoadPosts:         "Laden der Beiträge fehlgeschlagen",
		UsernameAlreadyTaken:      "Benutzername existiert schon",
		Content:                   "Schreibe hier deinen Inhalt! \nHier wird mit Markdown formatierter Text unterstützt!",
		CommentPlaceholder:        "Schreibe hier deinen Kommentar!",
		Comments:                  "Kommentare",
	},
}

func GetLanguageFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("Language")
	if err != nil || cookie.Value == "" {
		return "EN" // Default to English if cookie is not set
	}
	return cookie.Value
}

func GetTranslations(lang string) Translations {
	if t, ok := translations[lang]; ok {
		return t
	}
	return translations["EN"] // Fallback to English
}
