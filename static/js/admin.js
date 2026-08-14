// admin.js — frontend admin sobre el contrato REST JSON (/api/v1).
//
// Reemplaza el flujo HTMX (forms + claves naturales + HX-Redirect) por
// fetch + JSON + IDs numéricos + header X-CSRF-Token. El diseño visual NO
// cambia: cada modal deja de enviar forms por HTMX y las mutaciones llaman a
// estos helpers. Tras una mutación exitosa se recarga la página (location.
// reload): el server re-renderiza los datos nuevos.
//
// Contrato:
//   - La cookie `csrf_token` (NO HttpOnly) provee el token CSRF que todo
//     POST/PUT/DELETE admin manda en `X-CSRF-Token`.
//   - `admin_session` es HttpOnly: viaja sola con credentials 'same-origin'.
//   - Errores del server: {"error": "mensaje"} con el status HTTP.
//   - Creaciones: {"id": N}. 401 = sesión expirada (se recarga).
//
// Decisiones de v1 (documentadas en api_contract.md):
//   - Skill icon: NO hay endpoint de upload de icono (el contrato lo menciona
//     pero el server no lo implementa). El form lo mantiene pero no se sube.
//   - Project cover: NO hay endpoint de upload de cover. Se preserva el cover
//     existente al editar (data-cover) y el nuevo cover no se sube.
//   - Screenshot individual: el server no expone el imageId numérico de
//     project_images (solo URLs), por eso no hay botones de borrado por imagen.
//     La subida de screenshots (POST /projects/{id}/images) sí se usa.
(function () {
  'use strict';

  // ---------------------------------------------------------------------------
  // Helpers de cookies / DOM
  // ---------------------------------------------------------------------------

  function getCookie(name) {
    var m = document.cookie.match('(?:^|;)\\s*' + name + '=([^;]*)');
    return m ? decodeURIComponent(m[1]) : null;
  }

  // csrfToken lee la cookie csrf_token (no HttpOnly, legible por JS).
  function csrfToken() {
    return getCookie('csrf_token');
  }

  // showError pinta un mensaje en el div de error .admin-modal__error del modal.
  function showError(id, msg) {
    var el = document.getElementById(id);
    if (el) {
      el.textContent = msg || '';
    }
  }

  function value(id) {
    var el = document.getElementById(id);
    return el ? el.value : '';
  }

  function setValue(id, v) {
    var el = document.getElementById(id);
    if (el) {
      el.value = v === null || v === undefined ? '' : v;
    }
  }

  function boolValue(id) {
    return value(id) === 'true';
  }

  // csvValue convierte un input "a, b, c" en ["a","b","c"].
  function csvValue(id) {
    return value(id)
      .split(',')
      .map(function (t) { return t.trim(); })
      .filter(function (t) { return t !== ''; });
  }

  function num(s) {
    var n = parseInt(s, 10);
    return isNaN(n) ? 0 : n;
  }

  // nextPosition cuenta los items ya listados para ubicar un item nuevo al final.
  function nextPosition() {
    return document.querySelectorAll('.admin-skill-item').length;
  }

  function reload() {
    location.reload();
  }

  // ---------------------------------------------------------------------------
  // api() — wrapper fetch hacia /api/v1
  //   JSON:      Content-Type application/json + X-CSRF-Token
  //   multipart: NO Content-Type manual (el browser pone el boundary), solo X-CSRF-Token
  //   credentials 'same-origin' para que viajen admin_session y csrf_token.
  //   Status no-2xx -> rechaza con Error del {"error": ...}.
  //   401 -> la sesión expiró; el llamador decide (handleError recarga).
  //   2xx -> resuelve con el body JSON parseado (puede ser null).
  // ---------------------------------------------------------------------------
  function api(path, opts) {
    opts = opts || {};
    var headers = {};
    if (!opts.formData) {
      headers['Content-Type'] = 'application/json';
    }
    var token = csrfToken();
    if (token) {
      headers['X-CSRF-Token'] = token;
    }

    var init = {
      method: opts.method || 'GET',
      credentials: 'same-origin',
      headers: headers,
    };
    if (opts.formData) {
      init.body = opts.formData;
    } else if (opts.body !== undefined) {
      init.body = JSON.stringify(opts.body);
    }

    return fetch('/api/v1' + path, init).then(function (res) {
      return res.text().then(function (text) {
        var data = null;
        if (text) {
          try { data = JSON.parse(text); } catch (e) { /* body no JSON */ }
        }
        if (!res.ok) {
          var err = new Error((data && data.error) || 'Request failed (' + res.status + ')');
          err.status = res.status;
          err.data = data;
          throw err;
        }
        return data;
      });
    });
  }

  function upload(path, formData) {
    return api(path, { method: 'POST', formData: formData });
  }

  // uploadFiles sube un input file multiple (si tiene archivos) a path.
  function uploadFiles(fieldId, path, key) {
    var input = document.getElementById(fieldId);
    if (!input || !input.files || input.files.length === 0) {
      return Promise.resolve();
    }
    var fd = new FormData();
    for (var i = 0; i < input.files.length; i++) {
      fd.append(key, input.files[i], input.files[i].name);
    }
    return upload(path, fd);
  }

  function handleError(id, err) {
    showError(id, err.message || 'Something went wrong');
    if (err.status === 401) {
      setTimeout(reload, 1200);
    }
  }

  // ---------------------------------------------------------------------------
  // Login / Logout
  // ---------------------------------------------------------------------------

  function login(event) {
    event.preventDefault();
    var password = value('admin-password');
    if (!password) {
      showError('login-error', 'Please enter your password');
      return;
    }
    showError('login-error', '');
    api('/auth/login', { method: 'POST', body: { password: password } })
      .then(reload)
      .catch(function (err) {
        if (err.status === 429) {
          showError('login-error', 'Too many attempts, please try again in a minute.');
        } else {
          showError('login-error', 'Invalid password');
        }
      });
  }

  function logout() {
    api('/auth/logout', { method: 'POST' })
      .then(reload)
      .catch(function () { reload(); });
  }

  // ---------------------------------------------------------------------------
  // Skills
  // ---------------------------------------------------------------------------

  var editingSkillId = null;
  var editingSkillPosition = 0;

  function fillSkillForm(btn) {
    editingSkillId = num(btn.getAttribute('data-id'));
    editingSkillPosition = num(btn.getAttribute('data-position'));
    setValue('skill-name', btn.getAttribute('data-name'));
    setValue('skill-category', btn.getAttribute('data-is-tool') === 'true' ? 'true' : 'false');
    document.getElementById('skill-form-title').textContent = 'Edit Skill: ' + btn.getAttribute('data-name');
    document.getElementById('skill-submit-btn').textContent = 'Save Changes';
  }

  function resetSkillForm() {
    editingSkillId = null;
    editingSkillPosition = 0;
    setValue('skill-name', '');
    setValue('skill-category', 'false');
    setValue('skill-icon', '');
    document.getElementById('skill-form-title').textContent = 'Add New Skill';
    document.getElementById('skill-submit-btn').textContent = 'Add Skill';
    showError('skills-error', '');
  }

  function saveSkill(event) {
    event.preventDefault();
    showError('skills-error', '');
    var payload = {
      name: value('skill-name'),
      is_tool: boolValue('skill-category'),
      position: editingSkillId ? editingSkillPosition : nextPosition(),
    };
    var req = editingSkillId
      ? api('/skills/' + editingSkillId, { method: 'PUT', body: payload })
      : api('/skills', { method: 'POST', body: payload });
    req.then(reload).catch(function (err) { handleError('skills-error', err); });
  }

  function deleteSkill(btn) {
    var id = num(btn.getAttribute('data-id'));
    if (!id) {
      return;
    }
    var name = btn.getAttribute('data-name') || 'this skill';
    if (!confirm("Are you sure you want to delete '" + name + "'?")) {
      return;
    }
    api('/skills/' + id, { method: 'DELETE' })
      .then(reload)
      .catch(function (err) { handleError('skills-error', err); });
  }

  // ---------------------------------------------------------------------------
  // Projects
  // ---------------------------------------------------------------------------

  var editingProjectId = null;
  var editingProjectPosition = 0;
  var editingProjectCover = '';

  function fillProjectForm(btn) {
    editingProjectId = num(btn.getAttribute('data-id'));
    editingProjectPosition = num(btn.getAttribute('data-position'));
    editingProjectCover = btn.getAttribute('data-cover') || '';
    setValue('project-title-es', btn.getAttribute('data-title-es'));
    setValue('project-title-en', btn.getAttribute('data-title-en'));
    setValue('project-desc-es', btn.getAttribute('data-desc-es'));
    setValue('project-desc-en', btn.getAttribute('data-desc-en'));
    setValue('project-tech-desc-es', btn.getAttribute('data-tech-es'));
    setValue('project-tech-desc-en', btn.getAttribute('data-tech-en'));
    setValue('project-category', btn.getAttribute('data-category'));
    setValue('project-tags', btn.getAttribute('data-tags'));
    setValue('project-link', btn.getAttribute('data-link'));
    setValue('project-repo-link', btn.getAttribute('data-repo'));
    document.getElementById('project-form-title').textContent = 'Edit Project: ' + btn.getAttribute('data-title-en');
    document.getElementById('project-submit-btn').textContent = 'Save Changes';
  }

  function resetProjectForm() {
    editingProjectId = null;
    editingProjectPosition = 0;
    editingProjectCover = '';
    setValue('project-title-es', '');
    setValue('project-title-en', '');
    setValue('project-desc-es', '');
    setValue('project-desc-en', '');
    setValue('project-tech-desc-es', '');
    setValue('project-tech-desc-en', '');
    setValue('project-category', 'Web');
    setValue('project-tags', '');
    setValue('project-link', '');
    setValue('project-repo-link', '');
    setValue('project-cover', '');
    setValue('project-screenshots', '');
    document.getElementById('project-form-title').textContent = 'Add New Project';
    document.getElementById('project-submit-btn').textContent = 'Add Project';
    showError('projects-error', '');
  }

  function saveProject(event) {
    event.preventDefault();
    showError('projects-error', '');
    var payload = {
      title_es: value('project-title-es'),
      title_en: value('project-title-en'),
      description_es: value('project-desc-es'),
      description_en: value('project-desc-en'),
      tech_description_es: value('project-tech-desc-es'),
      tech_description_en: value('project-tech-desc-en'),
      category: value('project-category'),
      tags: csvValue('project-tags'),
      link: value('project-link'),
      repo_link: value('project-repo-link'),
      cover_url: editingProjectCover,
      position: editingProjectId ? editingProjectPosition : nextPosition(),
    };

    var promise;
    if (editingProjectId) {
      promise = api('/projects/' + editingProjectId, { method: 'PUT', body: payload })
        .then(function () {
          return uploadFiles('project-screenshots', '/projects/' + editingProjectId + '/images', 'screenshots');
        });
    } else {
      promise = api('/projects', { method: 'POST', body: payload })
        .then(function (data) {
          return uploadFiles('project-screenshots', '/projects/' + data.id + '/images', 'screenshots');
        });
    }
    promise.then(reload).catch(function (err) { handleError('projects-error', err); });
  }

  function deleteProject(btn) {
    var id = num(btn.getAttribute('data-id'));
    if (!id) {
      return;
    }
    var title = btn.getAttribute('data-title') || 'this project';
    if (!confirm("Are you sure you want to delete '" + title + "'?")) {
      return;
    }
    api('/projects/' + id, { method: 'DELETE' })
      .then(reload)
      .catch(function (err) { handleError('projects-error', err); });
  }

  // ---------------------------------------------------------------------------
  // Experience
  // ---------------------------------------------------------------------------

  var editingExperienceId = null;
  var editingExperiencePosition = 0;

  function fillExperienceForm(btn) {
    editingExperienceId = num(btn.getAttribute('data-id'));
    editingExperiencePosition = num(btn.getAttribute('data-position'));
    setValue('experience-company', btn.getAttribute('data-company'));
    setValue('experience-position-es', btn.getAttribute('data-position-es'));
    setValue('experience-position-en', btn.getAttribute('data-position-en'));
    setValue('experience-period-es', btn.getAttribute('data-period-es'));
    setValue('experience-period-en', btn.getAttribute('data-period-en'));
    setValue('experience-desc-es', btn.getAttribute('data-desc-es'));
    setValue('experience-desc-en', btn.getAttribute('data-desc-en'));
    document.getElementById('experience-form-title').textContent = 'Edit Experience: ' + btn.getAttribute('data-company');
    document.getElementById('experience-submit-btn').textContent = 'Save Changes';
  }

  function resetExperienceForm() {
    editingExperienceId = null;
    editingExperiencePosition = 0;
    setValue('experience-company', '');
    setValue('experience-position-es', '');
    setValue('experience-position-en', '');
    setValue('experience-period-es', '');
    setValue('experience-period-en', '');
    setValue('experience-desc-es', '');
    setValue('experience-desc-en', '');
    document.getElementById('experience-form-title').textContent = 'Add New Experience';
    document.getElementById('experience-submit-btn').textContent = 'Add Experience';
    showError('experience-error', '');
  }

  function saveExperience(event) {
    event.preventDefault();
    showError('experience-error', '');
    var payload = {
      company: value('experience-company'),
      position_es: value('experience-position-es'),
      position_en: value('experience-position-en'),
      period_es: value('experience-period-es'),
      period_en: value('experience-period-en'),
      description_es: value('experience-desc-es'),
      description_en: value('experience-desc-en'),
      position: editingExperienceId ? editingExperiencePosition : nextPosition(),
    };
    var req = editingExperienceId
      ? api('/experience/' + editingExperienceId, { method: 'PUT', body: payload })
      : api('/experience', { method: 'POST', body: payload });
    req.then(reload).catch(function (err) { handleError('experience-error', err); });
  }

  function deleteExperience(btn) {
    var id = num(btn.getAttribute('data-id'));
    if (!id) {
      return;
    }
    var company = btn.getAttribute('data-company') || '';
    var position = btn.getAttribute('data-position-en') || 'this role';
    if (!confirm("Are you sure you want to delete '" + position + "' role at '" + company + "'?")) {
      return;
    }
    api('/experience/' + id, { method: 'DELETE' })
      .then(reload)
      .catch(function (err) { handleError('experience-error', err); });
  }

  // ---------------------------------------------------------------------------
  // Education
  // ---------------------------------------------------------------------------

  var editingEducationId = null;
  var editingEducationPosition = 0;

  function fillEducationForm(btn) {
    editingEducationId = num(btn.getAttribute('data-id'));
    editingEducationPosition = num(btn.getAttribute('data-position'));
    setValue('education-title-es', btn.getAttribute('data-title-es'));
    setValue('education-title-en', btn.getAttribute('data-title-en'));
    setValue('education-school', btn.getAttribute('data-school'));
    setValue('education-date', btn.getAttribute('data-date'));
    setValue('education-desc-es', btn.getAttribute('data-desc-es'));
    setValue('education-desc-en', btn.getAttribute('data-desc-en'));
    setValue('education-category', btn.getAttribute('data-is-course') === 'true' ? 'true' : 'false');
    document.getElementById('education-form-title').textContent = 'Edit Education: ' + btn.getAttribute('data-title-en');
    document.getElementById('education-submit-btn').textContent = 'Save Changes';
  }

  function resetEducationForm() {
    editingEducationId = null;
    editingEducationPosition = 0;
    setValue('education-title-es', '');
    setValue('education-title-en', '');
    setValue('education-school', '');
    setValue('education-date', '');
    setValue('education-desc-es', '');
    setValue('education-desc-en', '');
    setValue('education-category', 'false');
    document.getElementById('education-form-title').textContent = 'Add New Entry';
    document.getElementById('education-submit-btn').textContent = 'Add Entry';
    showError('education-error', '');
  }

  function saveEducation(event) {
    event.preventDefault();
    showError('education-error', '');
    var payload = {
      title_es: value('education-title-es'),
      title_en: value('education-title-en'),
      school: value('education-school'),
      date: value('education-date'),
      is_course: boolValue('education-category'),
      description_es: value('education-desc-es'),
      description_en: value('education-desc-en'),
      position: editingEducationId ? editingEducationPosition : nextPosition(),
    };
    var req = editingEducationId
      ? api('/education/' + editingEducationId, { method: 'PUT', body: payload })
      : api('/education', { method: 'POST', body: payload });
    req.then(reload).catch(function (err) { handleError('education-error', err); });
  }

  function deleteEducation(btn) {
    var id = num(btn.getAttribute('data-id'));
    if (!id) {
      return;
    }
    var title = btn.getAttribute('data-title') || 'this entry';
    if (!confirm("Are you sure you want to delete '" + title + "'?")) {
      return;
    }
    api('/education/' + id, { method: 'DELETE' })
      .then(reload)
      .catch(function (err) { handleError('education-error', err); });
  }

  // ---------------------------------------------------------------------------
  // Hero / Profile
  // ---------------------------------------------------------------------------

  function saveHero(event) {
    event.preventDefault();
    showError('hero-error', '');
    var payload = {
      name: value('hero-name'),
      role_es: value('hero-role-es'),
      role_en: value('hero-role-en'),
      headline_es: value('hero-headline-es'),
      headline_en: value('hero-headline-en'),
      summary_es: value('hero-summary-es'),
      summary_en: value('hero-summary-en'),
    };
    api('/profile', { method: 'PUT', body: payload })
      .then(function () {
        return uploadFiles('hero-cv', '/profile/cv', 'cv');
      })
      .then(reload)
      .catch(function (err) { handleError('hero-error', err); });
  }

  // ---------------------------------------------------------------------------
  // Socials
  // ---------------------------------------------------------------------------

  function saveSocials(event) {
    event.preventDefault();
    showError('socials-error', '');
    var email = value('admin-email');
    var github = value('admin-github');
    var linkedin = value('admin-linkedin');

    var socials = [];
    if (github) {
      socials.push({ name: 'GitHub', url: github, icon_key: 'github', position: 0 });
    }
    if (linkedin) {
      socials.push({ name: 'LinkedIn', url: linkedin, icon_key: 'linkedin', position: 1 });
    }

    Promise.all([
      api('/profile', { method: 'PUT', body: { email: email } }),
      api('/socials', { method: 'PUT', body: socials }),
    ])
      .then(reload)
      .catch(function (err) { handleError('socials-error', err); });
  }

  // ---------------------------------------------------------------------------
  // API pública + funciones globales (referenciadas por los onclick inline)
  // ---------------------------------------------------------------------------

  window.PortfolioAdmin = {
    csrfToken: csrfToken,
    api: api,
    upload: upload,
    reload: reload,
    showError: showError,
    login: login,
    logout: logout,
    saveSkill: saveSkill,
    deleteSkill: deleteSkill,
    saveProject: saveProject,
    deleteProject: deleteProject,
    saveExperience: saveExperience,
    deleteExperience: deleteExperience,
    saveEducation: saveEducation,
    deleteEducation: deleteEducation,
    saveHero: saveHero,
    saveSocials: saveSocials,
  };

  // Funciones globales que los atributos onclick/onsubmit inline de los modales
  // referencian. Mantienen el contrato de los modales HTMX originales (nombres
  // idénticos), pero ahora operan con IDs numéricos.
  window.fillSkillForm = fillSkillForm;
  window.resetSkillForm = resetSkillForm;
  window.fillProjectForm = fillProjectForm;
  window.resetProjectForm = resetProjectForm;
  window.fillExperienceForm = fillExperienceForm;
  window.resetExperienceForm = resetExperienceForm;
  window.fillEducationForm = fillEducationForm;
  window.resetEducationForm = resetEducationForm;
})();
