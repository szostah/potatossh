package main

import (
	"fmt"
	"log"
	"net/http"
	"potatossh/internal/database"
	"potatossh/internal/session"
	"potatossh/internal/theme"
	"slices"
	"strconv"
	"text/template"
)

type Tab struct {
	Server  *database.ServerDbRow
	Session *session.Session
	Checked bool
}

type Window struct {
	Tabs   []Tab
	Active bool
}

type App struct {
	Db               *database.Database
	Servers          []*database.ServerOrDir
	Sessions         map[string]*session.Session
	Windows          []Window
	ActiveWindow     uint
	SessionWindowMap map[string]uint
	Template         *template.Template
	Themes           []theme.Theme
	Settings         database.Settings
}

func NewApp(dbFile string) *App {
	db, err := database.Open("potato.sqlite")
	if err != nil {
		log.Fatal(err)
	}

	themes := theme.Load()
	settings, err := db.GetSettings(database.Settings{Theme: &themes[0], FontSize: 10, OpenInNewWindow: false}, themes)
	if err != nil {
		log.Fatal(err)
	}

	app := &App{
		Db:               db,
		Servers:          []*database.ServerOrDir{},
		Sessions:         make(map[string]*session.Session),
		Windows:          []Window{},
		ActiveWindow:     0,
		SessionWindowMap: make(map[string]uint),
		Template:         template.New("index.html"),
		Themes:           themes,
		Settings:         settings,
	}
	app.Template, err = app.Template.Funcs(template.FuncMap{
		"openInNewWindowEnabled": func() bool {
			return app.Settings.OpenInNewWindow
		},
	}).ParseFiles("web/templates/index.html")

	if err != nil {
		log.Fatal(err)
	}

	return app
}

func (app *App) ToMap() map[string]any {
	return map[string]any{
		"Servers":  app.Servers,
		"Windows":  app.Windows,
		"Themes":   app.Themes,
		"Settings": app.Settings,
	}
}

func (app *App) UpdateServerList() error {
	list, err := app.Db.ServerListWithDirs()
	if err != nil {
		return err
	}
	app.Servers = list
	return nil
}

func (app *App) ServeHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	err := app.UpdateServerList()
	if err != nil {
		http.Error(w, "Database error", http.StatusBadRequest)
		return
	}
	app.Template.Execute(w, app.ToMap())
}

func (app *App) ServerRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		port, err := strconv.ParseUint(r.PostFormValue("port"), 10, 16)
		if err != nil {
			http.Error(w, "Can not parse port", http.StatusBadRequest)
			return
		}
		server := database.Server{Name: r.PostFormValue("name"), Address: r.PostFormValue("address"), Port: uint16(port), User: r.PostFormValue("user"), Password: r.PostFormValue("password")}
		_, err = app.Db.AddServer(&server)
		if err != nil {
			fmt.Println("Database error:", err)
			http.Error(w, "Database error", http.StatusBadRequest)
			return
		}
		fmt.Println("new server added")
	} else if r.Method == http.MethodDelete {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Can not parse id", http.StatusBadRequest)
			return
		}
		app.Db.DeleteServer(id)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	err := app.UpdateServerList()
	if err != nil {
		http.Error(w, "Database error", http.StatusBadRequest)
		return
	}
	app.Template.ExecuteTemplate(w, "server_list", app.Servers)
}

func (app *App) ConnectionRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		serverId, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Can not parse id", http.StatusBadRequest)
			return
		}
		server, err := app.Db.GetServer(serverId)
		if err != nil {
			http.Error(w, "Database error", http.StatusBadRequest)
			return
		}
		session, err := session.NewSession(server.Server, app.Template)
		if err != nil {
			http.Error(w, "Can not create the session.", http.StatusBadRequest)
			return
		}
		err = session.Start()
		if err != nil {
			http.Error(w, "Can not start the session.", http.StatusBadRequest)
			return
		}
		if r.PostFormValue("newwindow") == "true" {
			if len(app.Windows) > int(app.ActiveWindow) {
				app.Windows[app.ActiveWindow].Active = false
			}
			app.Windows = append(app.Windows, Window{Tabs: []Tab{{Server: &server, Session: session, Checked: true}}, Active: true})
			app.ActiveWindow = uint(len(app.Windows) - 1)
		} else {
			if len(app.Windows) == 0 {
				app.Windows = append(app.Windows, Window{Tabs: []Tab{{Server: &server, Session: session, Checked: true}}, Active: true})
				app.ActiveWindow = 0
			} else {
				app.Windows[app.ActiveWindow].Tabs = append(app.Windows[app.ActiveWindow].Tabs, Tab{Server: &server, Session: session, Checked: false})
			}
		}
		app.SessionWindowMap[session.Id] = app.ActiveWindow
		app.Sessions[session.Id] = session
		app.Template.ExecuteTemplate(w, "workspace", app.Windows)
	} else if r.Method == http.MethodGet {
		sessionId := r.PathValue("id")
		session, ok := app.Sessions[sessionId]
		if !ok {
			http.Error(w, "Requested session doesn't exist.", http.StatusBadRequest)
			return
		}
		err := session.AttachWebSocket(w, r)
		if err != nil {
			http.Error(w, "Can not attach to the session.", http.StatusBadRequest)
		}
	} else if r.Method == http.MethodDelete {
		sessionId := r.PathValue("id")
		session, ok := app.Sessions[sessionId]
		if !ok {
			http.Error(w, "Requested session doesn't exist.", http.StatusBadRequest)
			return
		}
		windowId := app.SessionWindowMap[sessionId]
		fmt.Println("SessionWindowMap", windowId)
		app.Windows[windowId].Tabs = slices.DeleteFunc(app.Windows[windowId].Tabs, func(tab Tab) bool {
			return tab.Session.Id == sessionId
		})

		if len(app.Windows[windowId].Tabs) == 0 {
			if windowId == app.ActiveWindow {
				if len(app.Windows) > 0 {
					app.Windows[0].Active = true
					app.ActiveWindow = 0
				}
			}
			app.RemoveWindow(windowId)
		} else {

		}

		session.Disconnect()
		delete(app.Sessions, sessionId)
		app.Template.ExecuteTemplate(w, "workspace", app.Windows)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) RemoveWindow(windowIdx uint) {
	app.Windows = append(app.Windows[:windowIdx], app.Windows[windowIdx+1:]...)

	for k, v := range app.SessionWindowMap {
		if v > windowIdx {
			app.SessionWindowMap[k] = v - 1
		}
	}

}

func (app *App) ValidateServerName(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.PostFormValue("name")
	unique, err := app.Db.IsNameUnique(name)
	if err != nil {
		http.Error(w, "Database error", http.StatusBadRequest)
		return
	}
	if unique {
		app.Template.ExecuteTemplate(w, "dialog_name", map[string]string{
			"Input": name,
		})
	} else {
		app.Template.ExecuteTemplate(w, "dialog_name", map[string]string{
			"Error": "This name is already used!",
			"Input": name,
		})
	}
}

func (app *App) SetActiveTab(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		sessionId := r.PathValue("id")
		fmt.Println("active tab", sessionId)
		for i, tab := range app.Windows[app.ActiveWindow].Tabs {
			if tab.Session.Id == sessionId {
				app.Windows[app.ActiveWindow].Tabs[i].Checked = true
			} else {
				app.Windows[app.ActiveWindow].Tabs[i].Checked = false
			}
		}
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) SetActiveWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		windowId, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Can not parse id", http.StatusBadRequest)
			return
		}
		app.Windows[app.ActiveWindow].Active = false
		app.Windows[windowId].Active = true
		app.ActiveWindow = uint(windowId)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) MoveTab(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		action := r.PathValue("action")
		sessionId := r.PathValue("sessionid")
		idx := slices.IndexFunc(app.Windows[app.ActiveWindow].Tabs, func(tab Tab) bool {
			return tab.Session.Id == sessionId
		})
		tab := app.Windows[app.ActiveWindow].Tabs[idx]
		if action == "newwindow" {
			if tab.Checked {
				if idx > 0 {
					app.Windows[app.ActiveWindow].Tabs[idx-1].Checked = true
				} else {
					app.Windows[app.ActiveWindow].Tabs[idx+1].Checked = true
				}
			}
			app.Windows[app.ActiveWindow].Active = false
			app.Windows[app.ActiveWindow].Tabs = append(app.Windows[app.ActiveWindow].Tabs[:idx], app.Windows[app.ActiveWindow].Tabs[idx+1:]...)
			tab.Checked = true
			new_window := Window{Tabs: []Tab{tab}, Active: true}
			app.Windows = append(app.Windows, new_window)
			app.SessionWindowMap[sessionId] = uint(len(app.Windows) - 1)
			app.ActiveWindow = uint(len(app.Windows) - 1)
		} else if action == "left" {
			if idx > 0 {
				app.Windows[app.ActiveWindow].Tabs[idx], app.Windows[app.ActiveWindow].Tabs[idx-1] = app.Windows[app.ActiveWindow].Tabs[idx-1], app.Windows[app.ActiveWindow].Tabs[idx]
			}
		} else if action == "right" {
			if idx < len(app.Windows[app.ActiveWindow].Tabs)-1 {
				app.Windows[app.ActiveWindow].Tabs[idx], app.Windows[app.ActiveWindow].Tabs[idx+1] = app.Windows[app.ActiveWindow].Tabs[idx+1], app.Windows[app.ActiveWindow].Tabs[idx]
			}
		} else {
			http.Error(w, "Move tab action not allowed", http.StatusBadRequest)
			return
		}
		app.Template.ExecuteTemplate(w, "workspace", app.Windows)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) SwitchWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		sessionId := r.PathValue("sessionid")
		windowId, err := strconv.Atoi(r.PathValue("windowid"))
		if err != nil {
			http.Error(w, "Can not parse id", http.StatusBadRequest)
			return
		}
		idx := slices.IndexFunc(app.Windows[app.ActiveWindow].Tabs, func(tab Tab) bool {
			return tab.Session.Id == sessionId
		})
		tab := app.Windows[app.ActiveWindow].Tabs[idx]
		if tab.Checked && len(app.Windows[app.ActiveWindow].Tabs) > 1 {
			if idx > 0 {
				app.Windows[app.ActiveWindow].Tabs[idx-1].Checked = true
			} else {
				app.Windows[app.ActiveWindow].Tabs[idx+1].Checked = true
			}
		}
		app.Windows[app.ActiveWindow].Active = false
		app.Windows[app.ActiveWindow].Tabs = append(app.Windows[app.ActiveWindow].Tabs[:idx], app.Windows[app.ActiveWindow].Tabs[idx+1:]...)
		tab.Checked = true
		app.Windows[windowId].Active = true
		app.Windows[windowId].Tabs = append(app.Windows[windowId].Tabs, tab)
		if len(app.Windows[app.ActiveWindow].Tabs) == 0 {
			app.RemoveWindow(app.ActiveWindow)
		}
		app.ActiveWindow = uint(windowId)
		app.Template.ExecuteTemplate(w, "workspace", app.Windows)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) ThemePreview(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if r.URL.Query().Has("theme") && r.URL.Query().Has("font_size") {
			theme_id, err := strconv.Atoi(r.URL.Query().Get("theme"))
			if err != nil {
				http.Error(w, "Can not parse theme id", http.StatusBadRequest)
				return
			}

			font_size, err := strconv.Atoi(r.URL.Query().Get("font_size"))
			if err != nil {
				http.Error(w, "Can not parse id", http.StatusBadRequest)
				return
			}
			preview := database.Settings{Theme: &app.Themes[theme_id], FontSize: uint(font_size)}
			app.Template.ExecuteTemplate(w, "settings_preview", preview)
		} else {
			app.Template.ExecuteTemplate(w, "settings_preview", app.Settings)
		}
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) ApplySettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		font_size, err := strconv.Atoi(r.PostFormValue("font_size"))
		if err != nil {
			http.Error(w, "Can not parse font_size", http.StatusBadRequest)
			return
		}

		fontsize_update := (app.Settings.FontSize != uint(font_size))
		app.Settings.FontSize = uint(font_size)

		theme_id, err := strconv.Atoi(r.PostFormValue("theme"))
		if err != nil {
			http.Error(w, "Can not parse theme id", http.StatusBadRequest)
			return
		}
		theme_update := (app.Settings.Theme != &app.Themes[theme_id])
		app.Settings.Theme = &app.Themes[theme_id]

		val, ok := r.PostForm["new_behavior"]
		openInNewWindow := ok && val[0] == "on"
		newbehavior_update := openInNewWindow != app.Settings.OpenInNewWindow
		app.Settings.OpenInNewWindow = openInNewWindow

		app.Template.ExecuteTemplate(w, "settings_form", app.ToMap())
		if theme_update {
			app.Template.ExecuteTemplate(w, "theme_oob", app.Settings.Theme)
		}
		if fontsize_update {
			app.Template.ExecuteTemplate(w, "fontsize_oob", app.Settings.FontSize)
		}
		if newbehavior_update {
			app.Template.ExecuteTemplate(w, "server_list_oob", app.Servers)
		}
		_, err = app.Db.UpdateSettings(&app.Settings)
		if err != nil {
			http.Error(w, "Database error!", http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (app *App) SetTitle(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		sessionId := r.PathValue("sessionid")
		session, ok := app.Sessions[sessionId]
		if !ok {
			http.Error(w, "Requested session doesn't exist.", http.StatusBadRequest)
			return
		}
		title, ok := r.Header["Hx-Prompt"]
		if !ok {
			http.Error(w, "Missing header.", http.StatusBadRequest)
			return
		}

		session.Terminal().SetStaticTitle(title[0])
		app.Template.ExecuteTemplate(w, "title_oob", session)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	app := NewApp("potato.sqlite")
	http.HandleFunc("/", app.ServeHome)
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web"+r.URL.Path)
	})
	http.HandleFunc("/server", app.ServerRequest)
	http.HandleFunc("/server/{id}", app.ServerRequest)
	http.HandleFunc("/validate/name", app.ValidateServerName)
	http.HandleFunc("/connection/{id}", app.ConnectionRequest)
	http.HandleFunc("/active/tab/{id}", app.SetActiveTab)
	http.HandleFunc("/active/window/{id}", app.SetActiveWindow)
	http.HandleFunc("/move/{action}/{sessionid}", app.MoveTab)
	http.HandleFunc("/move/window/{sessionid}/{windowid}", app.SwitchWindow)
	http.HandleFunc("/title/{sessionid}", app.SetTitle)
	http.HandleFunc("/preview", app.ThemePreview)
	http.HandleFunc("/settings", app.ApplySettings)

	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
