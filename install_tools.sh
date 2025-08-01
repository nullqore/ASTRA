#!/bin/bash

# This script installs all necessary tools for the SAAM project.

# Function to print messages
print_msg() {
    echo -e "\n[+] $1\n"
}

# --- System Dependencies ---
print_msg "Updating package lists and installing essential packages..."
sudo apt-get update
sudo apt-get install -y wget unzip git python3-pip jq

# --- Go Installation ---
if ! command -v go &> /dev/null
then
    print_msg "Go is not installed. Installing Go..."
    wget https://golang.org/dl/go1.24.5.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.24.5.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc
    source ~/.bashrc
    rm go1.24.5.linux-amd64.tar.gz
    print_msg "Go has been installed."
else
    print_msg "Go is already installed."
fi

# --- Go-based Tools ---
print_msg "Installing Go-based security tools..."
go install -v github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest
go install -v github.com/owasp-amass/amass/v5/cmd/amass@main
go install -v github.com/tomnomnom/assetfinder@latest
go install -v github.com/gwen001/github-subdomains@latest
go install -v github.com/projectdiscovery/chaos-client/cmd/chaos@latest
go install -v github.com/projectdiscovery/httpx/cmd/httpx@latest
go install -v github.com/projectdiscovery/naabu/v2/cmd/naabu@latest
go install -v github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest
go install -v github.com/projectdiscovery/katana/cmd/katana@latest
go install -v github.com/hakluke/hakrawler@latest
go install -v github.com/lc/gau@latest
go install -v github.com/projectdiscovery/urlfinder/cmd/urlfinder@latest
go install -v github.com/tomnomnom/waybackurls@latest
go install -v github.com/jaeles-project/gospider@latest
go install -v github.com/gwen001/github-endpoints@latest
go install -v github.com/trap-bytes/gourlex@latest
go install -v github.com/ffuf/ffuf@latest
go install -v github.com/tomnomnom/unfurl@latest
go install -v github.com/tomnomnom/qsreplace@latest
go install -v github.com/tomnomnom/anew@latest
go install -v github.com/hahwul/dalfox/v2@latest
go install -v github.com/shenwei356/rush@latest
go install -v github.com/deletescape/go-gauri@latest
go install -v github.com/003random/getJS/v2@latest
go install -v github.com/lc/subjs@latest
go install -v github.com/devanshbatham/arjun@latest
go install -v github.com/Brosck/mantra@latest
go install -v github.com/projectdiscovery/dnsx/cmd/dnsx@latest
go install -v github.com/ImAyrix/cut-cdn@latest
go install -v github.com/projectdiscovery/cdncheck/cmd/cdncheck@latest
go install -v github.com/nullqore/golinkfinder@latest

print_msg "Installing GF pattern..."
go install -v github.com/tomnomnom/gf@latest
git clone https://github.com/1ndianl33t/Gf-Patterns ~/tools/Gf-Patterns
mkdir -p ~/.gf
cp ~/tools/Gf-Patterns/*.json ~/.gf


# --- Python-based Tools ---
print_msg "Installing Python-based tools..."
sudo pip3 install sqlmap mantra
pip install waymore --break-system-packages

print_msg "Installing LinkFinder..."
git clone https://github.com/GerbenJavado/LinkFinder.git
cd LinkFinder
sudo python3 setup.py install
cd ..
rm -rf LinkFinder

print_msg "Installing LinkFinder..."
curl -LO https://github.com/findomain/findomain/releases/latest/download/findomain-linux.zip
unzip findomain-linux.zip
chmod +x findomain
sudo cp findomain /usr/local/bin/findomain && rm findomain-linux.zip

print_msg "Installing LinkFinder..."
git clone https://github.com/nsonaniya2010/SubDomainizer.git ~/tools/subdomainizer
pip3 install -r ~/tools/subdomainizer/requirements.txt --break-system-packages

# --- Other Tools ---
print_msg "Installing Aquatone..."
wget https://github.com/michenriksen/aquatone/releases/download/v1.7.0/aquatone_linux_amd64_1.7.0.zip
unzip aquatone_linux_amd64_1.7.0.zip
sudo mv aquatone /usr/local/bin/
rm aquatone_linux_amd64_1.7.0.zip

print_msg "Installing Paramspider..."
git clone https://github.com/devanshbatham/paramspider ~/tools/paramspider
cd ~/tools/paramspider && pip install . --break-system-packages && cd

print_msg "Installing Wappalyzer..."
sudo apt install pipx
pipx install wappalyzer
pipx ensurepath
print_msg "Installing subdominator..."
pipx install subdominator

wget https://raw.githubusercontent.com/orwagodfather/virustotalx/refs/heads/main/orwa.sh -O ~/tools/orwa.sh
chmod +x ~/tools/orwa.sh

# --- Finalizing ---
print_msg "Moving Go tools to /usr/local/bin..."
sudo cp ~/go/bin/* /usr/local/bin/

wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
sudo apt install ./google-chrome-stable_current_amd64.deb

git clone https://github.com/ameenmaali/urldedupe.git ~/tools/urldedupe
sudo snap install cmake --classic
cd ~/tools/urldedupe && cmake CMakeLists.txt && make && sudo cp urldedupe /usr/local/bin/ && cd 

print_msg "Installation complete. All tools are ready."
print_msg "You may need to configure the GF patterns by running 'gf -update' and setting up the patterns directory."
