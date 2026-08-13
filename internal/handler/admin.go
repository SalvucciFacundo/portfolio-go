package handler

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/SalvucciFacundo/portfolio-go/internal/auth"
	"github.com/SalvucciFacundo/portfolio-go/internal/data"
	"github.com/SalvucciFacundo/portfolio-go/internal/domain"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	password := r.FormValue("password")
	token, ok := auth.Login(password)
	if !ok {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`<div id="login-error" class="admin-modal__error">Invalid password</div>`))
		return
	}

	auth.SetSessionCookie(w, token)
	w.Header().Set("HX-Redirect", r.Referer())
	w.WriteHeader(http.StatusOK)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("admin_session"); err == nil {
		auth.Logout(cookie.Value)
	}
	auth.ClearSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HeroUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAdmin(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	profile := data.GetProfile()

	if v := r.FormValue("name"); v != "" {
		profile.Name = v
	}
	if v := r.FormValue("role_es"); v != "" {
		profile.RoleEs = v
	}
	if v := r.FormValue("role_en"); v != "" {
		profile.RoleEn = v
	}
	if v := r.FormValue("headline_es"); v != "" {
		profile.HeadlineEs = v
	}
	if v := r.FormValue("headline_en"); v != "" {
		profile.HeadlineEn = v
	}
	profile.SummaryEs = r.FormValue("summary_es")
	profile.SummaryEn = r.FormValue("summary_en")

	// Handle CV upload
	if file, header, err := r.FormFile("cv"); err == nil {
		defer file.Close()
		dst := filepath.Join("static", "uploads", header.Filename)
		_ = os.MkdirAll(filepath.Dir(dst), 0o755)
		out, err := os.Create(dst)
		if err == nil {
			defer out.Close()
			_, _ = io.Copy(out, file)
			profile.ResumeURL = "/static/uploads/" + header.Filename
		}
	}

	// Handle avatar upload
	if file, header, err := r.FormFile("avatar"); err == nil {
		defer file.Close()
		dst := filepath.Join("static", "uploads", header.Filename)
		_ = os.MkdirAll(filepath.Dir(dst), 0o755)
		out, err := os.Create(dst)
		if err == nil {
			defer out.Close()
			_, _ = io.Copy(out, file)
			profile.AvatarURL = "/static/uploads/" + header.Filename
		}
	}

	data.UpdateProfile(profile)

	w.Header().Set("HX-Redirect", r.Referer())
	w.WriteHeader(http.StatusOK)
}

func AvatarUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAdmin(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	dst := filepath.Join("static", "uploads", header.Filename)
	_ = os.MkdirAll(filepath.Dir(dst), 0o755)
	out, err := os.Create(dst)
	if err != nil {
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	_, _ = io.Copy(out, file)

	profile := data.GetProfile()
	profile.AvatarURL = "/static/uploads/" + header.Filename
	data.UpdateProfile(profile)

	w.Header().Set("HX-Redirect", r.Referer())
	w.WriteHeader(http.StatusOK)
}

func SkillsUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAdmin(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	oldName := r.FormValue("old_name")
	name := r.FormValue("name")
	isTool := r.FormValue("is_tool") == "true"

	profile := data.GetProfile()

	// Handle icon upload if any
	iconURL := ""
	if file, header, err := r.FormFile("icon"); err == nil {
		defer file.Close()
		dst := filepath.Join("static", "uploads", header.Filename)
		_ = os.MkdirAll(filepath.Dir(dst), 0o755)
		out, err := os.Create(dst)
		if err == nil {
			defer out.Close()
			_, _ = io.Copy(out, file)
			iconURL = "/static/uploads/" + header.Filename
		}
	}

	if oldName != "" {
		// Update existing skill
		for i, s := range profile.Skills {
			if s.Name == oldName {
				profile.Skills[i].Name = name
				profile.Skills[i].IsTool = isTool
				if iconURL != "" {
					profile.Skills[i].IconURL = iconURL
				}
				break
			}
		}
	} else {
		// Add new skill
		if iconURL == "" {
			iconURL = "https://placehold.co/64x64/1A1A1A/F4F4F2?text=" + name
		}
		newSkill := domain.Skill{
			Name:    name,
			IsTool:  isTool,
			IconURL: iconURL,
		}
		profile.Skills = append(profile.Skills, newSkill)
	}

	data.UpdateProfile(profile)

	w.Header().Set("HX-Redirect", r.Referer())
	w.WriteHeader(http.StatusOK)
}

func SkillsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAdmin(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "missing name parameter", http.StatusBadRequest)
		return
	}

	profile := data.GetProfile()
	newSkills := make([]domain.Skill, 0, len(profile.Skills))
	for _, s := range profile.Skills {
		if s.Name != name {
			newSkills = append(newSkills, s)
		}
	}
	profile.Skills = newSkills

	data.UpdateProfile(profile)

	w.Header().Set("HX-Redirect", r.Referer())
	w.WriteHeader(http.StatusOK)
}

func ProjectsUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAdmin(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	oldTitle := r.FormValue("old_title")
	titleEs := r.FormValue("title_es")
	titleEn := r.FormValue("title_en")
	descEs := r.FormValue("description_es")
	descEn := r.FormValue("description_en")
	techDescEs := r.FormValue("tech_description_es")
	techDescEn := r.FormValue("tech_description_en")
	category := r.FormValue("category")
	link := r.FormValue("link")
	repoLink := r.FormValue("repo_link")

	// Parse tags (technologies stack)
	var tags []string
	for _, t := range strings.Split(r.FormValue("tags"), ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}

	profile := data.GetProfile()

	// Handle single cover file upload
	coverURL := ""
	if file, header, err := r.FormFile("cover"); err == nil {
		defer file.Close()
		dst := filepath.Join("static", "uploads", header.Filename)
		_ = os.MkdirAll(filepath.Dir(dst), 0o755)
		out, err := os.Create(dst)
		if err == nil {
			defer out.Close()
			_, _ = io.Copy(out, file)
			coverURL = "/static/uploads/" + header.Filename
		}
	}

	// Handle multiple screenshots file upload
	var screenshotURLs []string
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		files := r.MultipartForm.File["screenshots"]
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err == nil {
				defer file.Close()
				dst := filepath.Join("static", "uploads", fileHeader.Filename)
				_ = os.MkdirAll(filepath.Dir(dst), 0o755)
				out, err := os.Create(dst)
				if err == nil {
					defer out.Close()
					_, _ = io.Copy(out, file)
					screenshotURLs = append(screenshotURLs, "/static/uploads/"+fileHeader.Filename)
				}
			}
		}
	}

	if oldTitle != "" {
		// Update existing project
		for i, pr := range profile.Projects {
			if pr.TitleEn == oldTitle {
				profile.Projects[i].TitleEs = titleEs
				profile.Projects[i].TitleEn = titleEn
				profile.Projects[i].DescriptionEs = descEs
				profile.Projects[i].DescriptionEn = descEn
				profile.Projects[i].TechDescriptionEs = techDescEs
				profile.Projects[i].TechDescriptionEn = techDescEn
				profile.Projects[i].Category = category
				profile.Projects[i].Tags = tags
				profile.Projects[i].Link = link
				profile.Projects[i].RepoLink = repoLink
				if coverURL != "" {
					profile.Projects[i].CoverURL = coverURL
				}
				if len(screenshotURLs) > 0 {
					profile.Projects[i].Screenshots = screenshotURLs
				} else if coverURL != "" {
					// Fallback to update first image of screenshots if cover changes
					profile.Projects[i].Screenshots = []string{coverURL}
				}
				break
			}
		}
	} else {
		// Add new project
		if coverURL == "" {
			coverURL = "https://placehold.co/800x500/1A1A1A/F4F4F2?text=" + strings.ReplaceAll(titleEn, " ", "+")
		}
		if len(screenshotURLs) == 0 {
			screenshotURLs = []string{coverURL}
		}
		newProj := domain.Project{
			TitleEs:           titleEs,
			TitleEn:           titleEn,
			DescriptionEs:     descEs,
			DescriptionEn:     descEn,
			TechDescriptionEs: techDescEs,
			TechDescriptionEn: techDescEn,
			Category:          category,
			Tags:              tags,
			Link:              link,
			RepoLink:          repoLink,
			CoverURL:          coverURL,
			Screenshots:       screenshotURLs,
		}
		profile.Projects = append(profile.Projects, newProj)
	}

	data.UpdateProfile(profile)

	w.Header().Set("HX-Redirect", r.Referer())
	w.WriteHeader(http.StatusOK)
}

func ProjectsDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAdmin(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	title := r.URL.Query().Get("title")
	if title == "" {
		http.Error(w, "missing title parameter", http.StatusBadRequest)
		return
	}

	profile := data.GetProfile()
	var newProjects []domain.Project
	for _, pr := range profile.Projects {
		if pr.TitleEn != title {
			newProjects = append(newProjects, pr)
		}
	}
	profile.Projects = newProjects

	data.UpdateProfile(profile)

	w.Header().Set("HX-Redirect", r.Referer())
	w.WriteHeader(http.StatusOK)
}

func EducationUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAdmin(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	oldTitle := r.FormValue("old_title")
	titleEs := r.FormValue("title_es")
	titleEn := r.FormValue("title_en")
	school := r.FormValue("school")
	date := r.FormValue("date")
	descEs := r.FormValue("description_es")
	descEn := r.FormValue("description_en")
	isCourse := r.FormValue("is_course") == "true"

	profile := data.GetProfile()

	if oldTitle != "" {
		// Update existing entry
		for i, e := range profile.Education {
			if e.TitleEn == oldTitle {
				profile.Education[i].TitleEs = titleEs
				profile.Education[i].TitleEn = titleEn
				profile.Education[i].School = school
				profile.Education[i].Date = date
				profile.Education[i].DescriptionEs = descEs
				profile.Education[i].DescriptionEn = descEn
				profile.Education[i].IsCourse = isCourse
				break
			}
		}
	} else {
		// Add new entry
		newEdu := domain.Education{
			TitleEs:       titleEs,
			TitleEn:       titleEn,
			School:        school,
			Date:          date,
			DescriptionEs: descEs,
			DescriptionEn: descEn,
			IsCourse:      isCourse,
		}
		profile.Education = append(profile.Education, newEdu)
	}

	data.UpdateProfile(profile)

	w.Header().Set("HX-Redirect", r.Referer())
	w.WriteHeader(http.StatusOK)
}

func EducationDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAdmin(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	title := r.URL.Query().Get("title")
	if title == "" {
		http.Error(w, "missing title parameter", http.StatusBadRequest)
		return
	}

	profile := data.GetProfile()
	var newEdu []domain.Education
	for _, e := range profile.Education {
		if e.TitleEn != title {
			newEdu = append(newEdu, e)
		}
	}
	profile.Education = newEdu

	data.UpdateProfile(profile)

	w.Header().Set("HX-Redirect", r.Referer())
	w.WriteHeader(http.StatusOK)
}

func ExperienceUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAdmin(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	oldCompany := r.FormValue("old_company")
	oldPosition := r.FormValue("old_position")
	company := r.FormValue("company")
	positionEs := r.FormValue("position_es")
	positionEn := r.FormValue("position_en")
	periodEs := r.FormValue("period_es")
	periodEn := r.FormValue("period_en")
	descEs := r.FormValue("description_es")
	descEn := r.FormValue("description_en")

	profile := data.GetProfile()

	if oldCompany != "" && oldPosition != "" {
		// Update existing entry
		for i, e := range profile.Experience {
			if e.Company == oldCompany && e.PositionEn == oldPosition {
				profile.Experience[i].Company = company
				profile.Experience[i].PositionEs = positionEs
				profile.Experience[i].PositionEn = positionEn
				profile.Experience[i].PeriodEs = periodEs
				profile.Experience[i].PeriodEn = periodEn
				profile.Experience[i].DescriptionEs = descEs
				profile.Experience[i].DescriptionEn = descEn
				break
			}
		}
	} else {
		// Add new entry
		newExp := domain.Experience{
			Company:       company,
			PositionEs:    positionEs,
			PositionEn:    positionEn,
			PeriodEs:      periodEs,
			PeriodEn:      periodEn,
			DescriptionEs: descEs,
			DescriptionEn: descEn,
		}
		profile.Experience = append(profile.Experience, newExp)
	}

	data.UpdateProfile(profile)

	w.Header().Set("HX-Redirect", r.Referer())
	w.WriteHeader(http.StatusOK)
}

func ExperienceDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAdmin(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	company := r.URL.Query().Get("company")
	position := r.URL.Query().Get("position")
	if company == "" || position == "" {
		http.Error(w, "missing company or position parameter", http.StatusBadRequest)
		return
	}

	profile := data.GetProfile()
	var newExp []domain.Experience
	for _, e := range profile.Experience {
		if e.Company != company || e.PositionEn != position {
			newExp = append(newExp, e)
		}
	}
	profile.Experience = newExp

	data.UpdateProfile(profile)

	w.Header().Set("HX-Redirect", r.Referer())
	w.WriteHeader(http.StatusOK)
}

func SocialsUpdateHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.IsAdmin(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	github := r.FormValue("github")
	linkedin := r.FormValue("linkedin")

	profile := data.GetProfile()
	profile.Email = email

	// Update GitHub link
	foundGithub := false
	for i, s := range profile.Socials {
		if s.IconKey == "github" {
			profile.Socials[i].URL = github
			foundGithub = true
			break
		}
	}
	if !foundGithub && github != "" {
		profile.Socials = append(profile.Socials, domain.SocialLink{
			Name:    "GitHub",
			URL:     github,
			IconKey: "github",
		})
	}

	// Update LinkedIn link
	foundLinkedin := false
	for i, s := range profile.Socials {
		if s.IconKey == "linkedin" {
			profile.Socials[i].URL = linkedin
			foundLinkedin = true
			break
		}
	}
	if !foundLinkedin && linkedin != "" {
		profile.Socials = append(profile.Socials, domain.SocialLink{
			Name:    "LinkedIn",
			URL:     linkedin,
			IconKey: "linkedin",
		})
	}

	data.UpdateProfile(profile)

	w.Header().Set("HX-Redirect", r.Referer())
	w.WriteHeader(http.StatusOK)
}



