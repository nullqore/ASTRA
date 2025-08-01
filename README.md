# ASTRA

<p align="center"><img src="https://github.com/nullqore/ASTRA/blob/main/media/intro.gif" width="100%" height="800"/></p>

---
> *_🚧Project under active development_*



**ASTRA (`Attack Surface Threat Reconnaissance Arsenal`) is an automated security assessment and management platform designed to streamline and orchestrate the entire reconnaissance and initial vulnerability analysis workflow.**

It provides a clean web-based user interface to manage security assessment projects, define scope, and run a series of powerful, open-source security tools in a structured and repeatable manner. The primary goal of ASTRA is to automate the tedious parts of a security assessment, allowing security professionals to focus on manual testing and deeper analysis.

## ❯ DEMO

| | |
|:---:|:---:|
| <strong>Index</strong><br><img src="https://github.com/nullqore/ASTRA/blob/main/media/pic1.png?raw=true" width="500" height="500" /> | <strong>Recon</strong><br><img src="https://github.com/nullqore/ASTRA/blob/main/media/pic2.png?raw=true" width="500" height="500" /> |
| <strong>Attack</strong><br><img src="https://github.com/nullqore/ASTRA/blob/main/media/pic3.png?raw=true" width="500" height="500" /> | <strong>Report</strong><br><img src="https://github.com/nullqore/ASTRA/blob/main/media/pic4.png?raw=true" width="500" height="500" /> |

---

## ❯ Features

- **Intuitive Web Interface:** Manage projects and targets through a modern and responsive frontend.
- **Modular Reconnaissance:** Select and run from a wide array of reconnaissance modules.
- **Live Task Monitoring:** View the real-time output of running scans directly in the browser via WebSockets.
- **Structured Workflow:** A logical, multi-stage process from subdomain discovery to vulnerability scanning.
- **Organized Results:** All tool outputs are systematically saved and organized by project, making data correlation and review straightforward.
- **Extensible Architecture:** Easily add new modules and tools to the backend.

---

## ❯ Architecture

ASTRA is built with a simple yet powerful client-server architecture:

-   **Backend:** A Go (Golang) server that exposes a REST API for project management and a WebSocket endpoint for live recon output. It acts as the orchestrator, managing the execution of various command-line security tools based on the selected modules.
-   **Frontend:** A vanilla HTML, JavaScript, and Tailwind CSS single-page application that provides the user interface for interacting with the backend.
-   **Tooling:** The platform integrates a curated list of best-in-class, open-source security tools to perform the actual scanning and analysis.

---

## ❯ Workflow

The typical workflow for conducting a security assessment with ASTRA is as follows:

1.  **Project Creation:** Start by creating a new project, giving it a unique name.
2.  **Scope Definition:** Define the project's scope by adding target domains, wildcards, and out-of-scope targets.
3.  **Module Selection:** Navigate to the "Deep Recon" page and select the reconnaissance modules you wish to run for the current project.
4.  **Execution & Monitoring:** Start the reconnaissance task. The backend will begin executing the selected modules in a logical sequence. You can monitor the progress and see the live output from the tools directly in the web UI.
5.  **Review Results:** Once the scan is complete, all artifacts and tool outputs are stored in the `results/<project-name>/` directory. You can then analyze this data for potential vulnerabilities.

---

## ❯ Reconnaissance Modules & Tools

ASTRA's power comes from its modular workflow, where each module is responsible for a specific part of the reconnaissance process and utilizes one or more specialized tools.

| Module | Purpose | Tools Used |
| :--- | :--- | :--- |
| `subdomain_discovery` | Discovers subdomains for the target scope. | `subfinder`, `amass`, `assetfinder`, `chaos` |
| `probe` | Checks which of the discovered subdomains are live and accessible via HTTP/HTTPS. | `httpx` |
| `portscan` | Scans for open ports on the discovered subdomains. | `naabu` |
| `urls_crawler` | Crawls live websites to find URLs and endpoints. | `katana`, `hakrawler`, `gau` |
| `js` | Finds and analyzes JavaScript files for secrets, endpoints, and vulnerabilities. | `getjs`, `subjs`, `nuclei` (exposures), `mantra`, `linkfinder` |
| `tech_detect` | Identifies the technologies used by the target web applications. | `wappalyzer` (via `httpx`) |
| `hidden_parameter` | Discovers hidden parameters in URLs that could be vulnerable. | `gf` (various patterns), `unfurl`, `arjun` |
| `fuzzer` | Fuzzes web applications for hidden files, directories, and vhosts. | `ffuf` |
| `vuln_scan` | Performs broad vulnerability scanning using Nuclei templates. | `nuclei` |
| `xss_scan` | Scans for Cross-Site Scripting (XSS) vulnerabilities. | `qsreplace`, `dalfox`, `ffuf` |
| `sqli_scan` | Scans for SQL Injection (SQLi) vulnerabilities. | `sqlmap`, `gauri` |
| `screenshot` | Takes screenshots of live web applications for visual inspection. | `aquatone` |

---

## ❯ Installation & Setup

Follow these steps to get your ASTRA instance up and running.

### 1. Prerequisites

Ensure you have the following installed on your system:
- `git`
- `go` (version 1.17+)
- `python3` and `pip`
- `npm`

### 2. Clone the Repository

```bash
git clone https://github.com/nullqore/ASTRA.git project-astra
cd project-astra

```

### 3. Install Tools

The project includes a comprehensive setup script to install all the required security tools. Make it executable and run it.

```bash
chmod +x install_tools.sh
sudo ./install_tools.sh
```

**Note:** This script will install numerous packages and binaries onto your system using `apt`, `go install`, `pip`, and `npm`. Review the script if you have any concerns.

### 4. Build and Run the Backend

Compile and run the Go server.

```bash
# Navigate to the backend directory
cd backend

# Build the binary
go build -o bin/server cmd/server/main.go
# Run the server
./bin/server

or

go mod init
go run cmd/server/main.go

```

The backend server will start on `localhost:8080` by default.

### 5. Access the Frontend

No web server is needed for the frontend. Simply open the `index.html` file in your browser.

```bash
# From the project root directory
open frontend/index.html
# Or navigate to file:///path/to/project-astra/frontend/index.html in your browser
```

---
```bash
ps aux | grep toolname //check tool running status
```
---
## ❯ Results Structure

All output from the reconnaissance process is saved within the `results/` directory, neatly organized by project and module:

```
results/
└── <project_name>/
    ├── scope/
    ├── subs/
    ├── httpx/
    ├── active/
    ├── urls/
    ├── portscan/
    ├── jsurls/
    └── vuln/
```

---
## ❯ Disclaimer

ASTRA is a tool intended for authorized security assessments and educational purposes only. Unauthorized scanning of networks and systems is illegal. The developers assume no liability and are not responsible for any misuse or damage caused by this program.
