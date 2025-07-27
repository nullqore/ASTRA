document.addEventListener('DOMContentLoaded', () => {
    console.log("SAAM Script Initializing");

    const API_BASE_URL = 'http://localhost:8080';
    const mainView = document.getElementById('main-view');
    const dashboardView = document.getElementById('dashboard-view');
    const createProjectForm = document.getElementById('create-project-form');
    const projectNameInput = document.getElementById('project-name');
    const projectListContainer = document.getElementById('project-list-container');
    const dashboardTitle = document.getElementById('dashboard-title');
    const backToProjectsBtn = document.getElementById('back-to-projects-btn');
    const enterBtn = document.getElementById('enter-btn');
    const projectsSection = document.getElementById('projects');

    async function fetchApi(url, options = {}) {
        try {
            const response = await fetch(url, options);
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({ error: `Request failed: ${response.statusText}` }));
                throw new Error(errorData.error);
            }
            return response.json();
        } catch (error) {
            console.error(`API call to ${url} failed:`, error);
            showNotification(`API call to ${url} failed`, true);
            throw error;
        }
    }

    const showView = (viewName) => {
        mainView.style.display = 'none';
        dashboardView.style.display = 'none';
        if (viewName === 'main') mainView.style.display = 'block';
        else if (viewName === 'dashboard') dashboardView.style.display = 'block';
    };

    function renderDashboard(project) {
        if (!project || !project.name) {
            showMainView();
            return;
        }
        localStorage.setItem('activeProjectName', project.name);
        window.location.href = 'dashboard.html';
    }

    async function showDashboardByName(projectName) {
        try {
            const projectDetails = await fetchApi(`${API_BASE_URL}/api/projects/${projectName}`);
            renderDashboard(projectDetails);
        } catch (error) {
            showNotification(`Failed to load dashboard for project: ${projectName}`, true);
            showMainView();
        }
    }

    function showMainView() {
        localStorage.removeItem('activeProjectName');
        showView('main');
        fetchAndRenderProjects();
    }
    
    async function fetchAndRenderProjects() {
        try {
            const projects = await fetchApi(`${API_BASE_URL}/api/projects`);
            const projectListEl = document.getElementById('project-list');
            projectListEl.innerHTML = ''; 

            if (!projects || projects.length === 0) {
                projectListEl.innerHTML = `<p class="text-center text-gray-500">No active projects found.</p>`;
            } else {
                projects.forEach(p => {
                    const projectEl = document.createElement('div');
                    projectEl.className = 'project-item';
                    projectEl.dataset.projectName = p.name;
                    projectEl.innerHTML = `<h3>${p.name}</h3>`;
                    projectListEl.appendChild(projectEl);
                });
            }
        } catch (error) { 
            console.error("Failed to render projects:", error);
            showNotification("Failed to load project list.", true);
        }
    }

    enterBtn.addEventListener('click', () => {
        projectsSection.style.display = 'block';
        projectsSection.scrollIntoView({ behavior: 'smooth' });
    });

    projectListContainer.addEventListener('click', (e) => {
        const projectItem = e.target.closest('.project-item');
        if (projectItem && projectItem.dataset.projectName) {
            showDashboardByName(projectItem.dataset.projectName);
        }
    });

    createProjectForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const projectName = projectNameInput.value.trim();
        if (!projectName) {
            showNotification('Project name cannot be empty.', true);
            return;
        }
    
        try {
            const newProject = await fetchApi(`${API_BASE_URL}/api/create-project`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ projectName })
            });
            
            projectNameInput.value = '';
            showNotification('Project created successfully!');
            renderDashboard(newProject);
        } catch (error) {
            showNotification(error.message, true);
            console.error('Error:', error);
        }
    });

    if (backToProjectsBtn) {
        backToProjectsBtn.addEventListener('click', showMainView);
    }

    function initializeApp() {
        const activeProjectName = localStorage.getItem('activeProjectName');
        if (activeProjectName && window.location.pathname.includes('dashboard.html')) {
            showDashboardByName(activeProjectName);
        } else {
            showMainView();
        }
    }

    initializeApp();
});

function showNotification(message, isError = false) {
    const notification = document.getElementById('notification');
    if (!notification) return;
    notification.textContent = message;
    notification.className = `notification show ${isError ? 'error' : ''}`;
    setTimeout(() => {
        notification.className = 'notification';
    }, 3000);
}