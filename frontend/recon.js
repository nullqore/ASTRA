document.addEventListener('DOMContentLoaded', () => {
    // --- 3D Background Initialization ---
    function initThree() {
        if (typeof THREE === 'undefined') return;
        const scene = new THREE.Scene();
        const camera = new THREE.PerspectiveCamera(75, window.innerWidth / window.innerHeight, 1, 10000);
        camera.position.z = 1000;
        const renderer = new THREE.WebGLRenderer({ canvas: document.getElementById('three-canvas'), alpha: true });
        renderer.setSize(window.innerWidth, window.innerHeight);
        const particleCount = 5000;
        const geometry = new THREE.BufferGeometry();
        const positions = new Float32Array(particleCount * 3);
        for (let i = 0; i < particleCount; i++) {
            positions[i * 3] = Math.random() * 2000 - 1000;
            positions[i * 3 + 1] = Math.random() * 2000 - 1000;
            positions[i * 3 + 2] = Math.random() * 2000 - 1000;
        }
        geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
        const material = new THREE.PointsMaterial({ size: 2, color: 0x00c3ff, blending: THREE.AdditiveBlending, transparent: true, opacity: 0.7 });
        const particles = new THREE.Points(geometry, material);
        scene.add(particles);
        let mouseX = 0;
        document.addEventListener('mousemove', (event) => { mouseX = event.clientX - (window.innerWidth / 2); }, false);
        window.addEventListener('resize', () => {
            camera.aspect = window.innerWidth / window.innerHeight;
            camera.updateProjectionMatrix();
            renderer.setSize(window.innerWidth, window.innerHeight);
        }, false);
        function animate() {
            requestAnimationFrame(animate);
            camera.position.x += (mouseX - camera.position.x) * 0.05;
            camera.lookAt(scene.position);
            particles.rotation.y = Date.now() * 0.00005;
            renderer.render(scene, camera);
        }
        animate();
    }
    initThree();

    // --- Recon Page Logic ---
    const projectName = localStorage.getItem('activeProjectName');
    if (projectName) {
        document.getElementById('recon-project-name').textContent = `Project: ${projectName}`;
    } else {
        window.location.href = 'index.html';
        return;
    }

    const toolListContainer = document.getElementById('recon-tool-list');
    const startReconBtn = document.getElementById('start-recon-btn');
    const pauseReconBtn = document.getElementById('pause-recon-btn');
    const stopReconBtn = document.getElementById('stop-recon-btn');
    const liveOutput = document.getElementById('live-output');
    const progressOutput = document.getElementById('progress-output');
    const backToProjectsBtn = document.getElementById('back-to-projects-btn');
    let socket;
    let isPaused = false;

    const reconModules = [
        "subfinder", "probe", "port_scan", "urls_crawler",
        "js_crawler", "tech_detect", "paramspyder",
        "fuzzer", "screenshot", "vuln_scan", "xss_scan", "sqli_scan"
    ];

    function renderModules() {
        toolListContainer.innerHTML = '';
        reconModules.forEach(tool => {
            const label = document.createElement('label');
            label.className = 'tool-checkbox-label';
            label.innerHTML = `
                <input type="checkbox" class="tool-checkbox" data-tool-name="${tool}">
                <span>${tool.replace(/_/g, ' ')}</span>
            `;
            toolListContainer.appendChild(label);
        });
    }

    function connectWebSocket() {
        socket = new WebSocket('ws://localhost:8080/ws');

        socket.onopen = () => {
            console.log('WebSocket connection established.');
            liveOutput.innerHTML = '> Connection established with server.\n';
            socket.send(JSON.stringify({ action: 'status', project: projectName }));
        };

        socket.onmessage = (event) => {
            const data = JSON.parse(event.data);

            if (data.log) {
                liveOutput.innerHTML = data.log.replace(/\n/g, '<br>');
                liveOutput.scrollTop = liveOutput.scrollHeight;
            }

            if (data.progress) {
                progressOutput.textContent = data.progress.substring(1);
            } else if (data.log && data.log.includes('Probed')) {
                progressOutput.textContent = data.log;
            }

            if (data.status) {
                updateUIForStatus(data.status);
            }
        };

        socket.onclose = () => {
            console.log('WebSocket connection closed.');
            liveOutput.innerHTML += '<br>> Connection closed. Please refresh the page to reconnect.';
            updateUIForStatus('stopped');
        };

        socket.onerror = (error) => {
            console.error('WebSocket error:', error);
            liveOutput.innerHTML += `<br><span class="text-red-500">> WebSocket error. See console for details.</span>`;
        };
    }

    function updateUIForStatus(status) {
        if (status === 'running') {
            startReconBtn.classList.add('hidden');
            pauseReconBtn.classList.remove('hidden');
            stopReconBtn.classList.remove('hidden');
            isPaused = false;
            pauseReconBtn.classList.remove('text-blue-400'); // Remove color when running
            pauseReconBtn.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-pause"><rect x="6" y="4" width="4" height="16"></rect><rect x="14" y="4" width="4" height="16"></rect></svg>';
        } else if (status === 'paused') {
            startReconBtn.classList.add('hidden');
            pauseReconBtn.classList.remove('hidden');
            stopReconBtn.classList.remove('hidden');
            isPaused = true;
            pauseReconBtn.classList.add('text-blue-400');
            pauseReconBtn.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-play"><polygon points="5 3 19 12 5 21 5 3"></polygon></svg>';
        } else {
            startReconBtn.classList.remove('hidden');
            pauseReconBtn.classList.add('hidden');
            stopReconBtn.classList.add('hidden');
            const selectedTools = toolListContainer.querySelectorAll('.tool-checkbox:checked');
            startReconBtn.disabled = selectedTools.length === 0;
        }
    }

    renderModules();
    connectWebSocket();

    toolListContainer.addEventListener('change', () => {
        const selectedTools = toolListContainer.querySelectorAll('.tool-checkbox:checked');
        startReconBtn.disabled = selectedTools.length === 0 || !socket || socket.readyState !== WebSocket.OPEN;
    });

    startReconBtn.addEventListener('click', () => {
        const selectedTools = Array.from(toolListContainer.querySelectorAll('.tool-checkbox:checked'))
            .map(checkbox => checkbox.dataset.toolName);

        if (selectedTools.length > 0 && socket && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify({ action: 'start', project: projectName, modules: selectedTools }));
        }
    });

    pauseReconBtn.addEventListener('click', () => {
        const action = isPaused ? 'resume' : 'pause';
        socket.send(JSON.stringify({ action: action, project: projectName }));
    });

    stopReconBtn.addEventListener('click', () => {
        socket.send(JSON.stringify({ action: 'stop', project: projectName }));
    });

    backToProjectsBtn.addEventListener('click', () => {
        localStorage.removeItem('activeProjectName');
        window.location.href = 'index.html';
    });
});
