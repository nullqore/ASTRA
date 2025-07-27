function initThree() {
    if (typeof THREE === 'undefined') { console.error("THREE.js not loaded."); return; }
    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(75, window.innerWidth / window.innerHeight, 1, 10000);
    camera.position.z = 1000;
    const renderer = new THREE.WebGLRenderer({ canvas: document.getElementById('three-canvas'), alpha: true });
    renderer.setSize(window.innerWidth, window.innerHeight);
    const particles = new THREE.Points(new THREE.BufferGeometry(), new THREE.PointsMaterial({ size: 2, color: 0x00c3ff, blending: THREE.AdditiveBlending, transparent: true, opacity: 0.7 }));
    const positions = new Float32Array(5000 * 3);
    for (let i = 0; i < positions.length; i++) {
        positions[i] = Math.random() * 2000 - 1000;
    }
    particles.geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    scene.add(particles);
    let mouseX = 0;
    document.addEventListener('mousemove', (e) => { mouseX = e.clientX - window.innerWidth / 2; });
    window.addEventListener('resize', () => {
        camera.aspect = window.innerWidth / window.innerHeight;
        camera.updateProjectionMatrix();
        renderer.setSize(window.innerWidth, window.innerHeight);
    });
    function animate() {
        requestAnimationFrame(animate);
        camera.position.x += (mouseX - camera.position.x) * 0.05;
        camera.lookAt(scene.position);
        particles.rotation.y = Date.now() * 0.00005;
        renderer.render(scene, camera);
    }
    animate();
    console.log("Three.js background initialized successfully.");
}

document.addEventListener('DOMContentLoaded', initThree);
