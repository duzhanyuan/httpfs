package webapi

import (
	"encoding/json"
	"fmt"
	"io"
	//"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/prologic/httpfs/utils"
)

func toHTTPError(err error) (msg string, httpStatus int) {
	switch {
	case os.IsPermission(err):
		return "Forbidden", http.StatusForbidden
	case os.IsNotExist(err):
		return "File Not Found", http.StatusNotFound
	case os.IsExist(err):
		return "File Already Exists", http.StatusConflict
	default:
		return "Internal Server Error", http.StatusInternalServerError
	}
}

func addStatHeaders(w http.ResponseWriter, stat os.FileInfo) {
	if w.Header().Get("Content-Length") == "" {
		w.Header().Set(
			"Content-Length",
			fmt.Sprintf("%d", stat.Size()),
		)
	}

	if w.Header().Get("Last-Modified") == "" {
		w.Header().Set(
			"Last-Modified",
			stat.ModTime().UTC().Format(http.TimeFormat),
		)
	}

	if w.Header().Get("X-File-Mode") == "" {
		w.Header().Set(
			"X-File-Mode",
			fmt.Sprintf("%d", stat.Mode()),
		)
	}

	if w.Header().Get("X-Is-Dir") == "" {
		w.Header().Set(
			"X-Is-Dir",
			fmt.Sprintf("%t", stat.IsDir()),
		)
	}
}

// FileServer ...
func FileServer(dir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		urlPath := path.Clean(r.URL.Path)
		localPath := path.Join(dir, urlPath)
		switch r.Method {
		case "HEAD":
			d, err := os.Lstat(localPath)
			if err != nil {
				//log.Printf("E: os.Lstat('%s') -> %s\n", localPath, err)
				msg, code := toHTTPError(err)
				http.Error(w, msg, code)
				return
			}

			addStatHeaders(w, d)

			return
		case "DELETE":
			err := os.RemoveAll(localPath)
			if err != nil {
				//log.Printf("E: os.RemoveAll('%s') -> %s\n", localPath, err)
				msg, code := toHTTPError(err)
				http.Error(w, msg, code)
				return
			}

			return
		case "PUT":
			query := r.URL.Query()

			perm := os.FileMode(
				utils.SafeParseInt(
					query.Get("perm"),
					0666,
				),
			)

			flags := utils.SafeParseInt(
				query.Get("flags"),
				os.O_WRONLY|os.O_CREATE|os.O_EXCL,
			)

			offset := utils.SafeParseInt64(
				query.Get("offset"),
				0,
			)

			//size := utils.SafeStatSize(localPath)

			//log.Printf(" size=%d\n", size)
			//log.Printf(" flags=%d\n", flags)
			//log.Printf(" offset=%d\n", offset)

			f, err := os.OpenFile(localPath, flags, perm)
			defer f.Close()
			if err != nil {
				//log.Printf("E: os.OpenFile('%s') -> %s\n", localPath, err)
				msg, code := toHTTPError(err)
				http.Error(w, msg, code)
				return
			}

			SeekType := io.SeekStart

			if offset < 0 {
				offset = -offset
				SeekType = io.SeekEnd
			}

			//log.Printf(" seeking to %d\n", offset)
			f.Seek(offset, SeekType)

			cl := utils.SafeParseInt64(r.Header.Get("Content-Length"), 0)

			n, err := io.Copy(f, r.Body)
			if err != nil {
				//log.Printf("E: io.Copy(...) -> %s\n", localPath, err)
				msg, code := toHTTPError(err)
				http.Error(w, msg, code)
				return
			}

			if cl != 0 && n == cl {
				return
			}

			w.Write([]byte(fmt.Sprintf("%d", n)))

			http.Error(w, "Partial Content", http.StatusPartialContent)
			return
		case "GET":
			d, err := os.Stat(localPath)
			if err != nil {
				//log.Printf("E: os.Stat('%s') -> %s\n", localPath, err)
				msg, code := toHTTPError(err)
				http.Error(w, msg, code)
				return
			}

			if d.IsDir() && !strings.HasSuffix(r.URL.Path, "/") {
				http.Redirect(w, r, r.URL.Path+"/", 302)
				return
			}

			if d.IsDir() {
				entries, err := utils.ReadDir(localPath)
				if err != nil {
					//log.Printf("E: utils.ReadDir('%s') -> %s\n", localPath, err)
					msg, code := toHTTPError(err)
					http.Error(w, msg, code)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(entries)
			} else {
				f, err := os.Open(localPath)
				if err != nil {
					//log.Printf("E: os.Open('%s') -> %s\n", localPath, err)
					msg, code := toHTTPError(err)
					http.Error(w, msg, code)
					return
				}
				defer f.Close()
				//log.Printf("Serving: %s\n", localPath)

				addStatHeaders(w, d)
				http.ServeContent(w, r, d.Name(), d.ModTime(), f)
			}
			return
		case "CHMOD":
			mode := utils.SafeParseInt(r.URL.Query().Get("mode"), 0)

			err := os.Chmod(localPath, os.FileMode(mode))
			if err != nil {
				//log.Printf("E: os.Chmod('%s', %d) -> %s\n", localPath, mode, err,)
				msg, code := toHTTPError(err)
				http.Error(w, msg, code)
				return
			}

			return
		case "MKDIR":
			perm := os.FileMode(
				utils.SafeParseInt(
					r.URL.Query().Get("perm"),
					0777,
				),
			)

			err := os.Mkdir(localPath, perm)

			if err != nil {
				//log.Printf("E: os.Mkdir(%q) -> %s\n", localPath, err)
				msg, code := toHTTPError(err)
				http.Error(w, msg, code)
				return
			}

			return
		case "LINK":
			nameReq := r.URL.Query().Get("name")
			if nameReq == "" {
				//log.Printf(" E: No ?name= specified for LINK request\n")
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			nameReq, err := url.QueryUnescape(nameReq)
			if err != nil {
				//log.Printf("E: %s\n", err)
				msg, code := toHTTPError(err)
				http.Error(w, msg, code)
				return
			}

			soft := utils.SafeParseBool(r.URL.Query().Get("soft"), false)

			namePath := path.Clean(nameReq)
			toPath := path.Join(dir, namePath)

			//log.Printf(" name=%q\n", nameReq)
			//log.Printf(" localPath=%q\n", localPath)
			//log.Printf(" toPath=%q\n", toPath)

			if soft {
				err = os.Symlink(localPath, toPath)
			} else {
				err = os.Link(localPath, toPath)
			}

			if err != nil {
				//log.Printf( "E: os.Link(%q, %q) -> %s\n", localPath, toPath, err,)
				msg, code := toHTTPError(err)
				http.Error(w, msg, code)
				return
			}

			return
		case "RENAME":
			nameReq := r.URL.Query().Get("name")
			if nameReq == "" {
				//log.Printf("E: No ?name= specified for RENAME request\n")
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			namePath := path.Clean(nameReq)
			toPath := path.Join(dir, namePath)

			err := os.Rename(localPath, toPath)
			if err != nil {
				//log.Printf( "E: os.Rename('%s', '%s') -> %s\n", localPath, toPath, err,)
				msg, code := toHTTPError(err)
				http.Error(w, msg, code)
				return
			}

			return
		case "TRUNCATE":
			sizeReq := r.URL.Query().Get("size")
			if sizeReq == "" {
				//log.Printf("E: No ?size= specified for TRUNCATE request\n")
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			size, err := strconv.ParseInt(sizeReq, 10, 32)
			if err != nil {
				//log.Printf( "E: strconv.ParseInt('%s', 10, 32) -> %s\n", sizeReq, err,)
				msg, code := toHTTPError(err)
				http.Error(w, msg, code)
				return
			}

			err = os.Truncate(localPath, size)
			if err != nil {
				//log.Printf( "E: os.Truncate('%s', %d) -> %s\n", localPath, size, err,)
				msg, code := toHTTPError(err)
				http.Error(w, msg, code)
				return
			}

			return
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		return
	}
}
