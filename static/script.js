document.getElementById('folderInput').addEventListener('change', function(e) {
    const folderName = e.target.files[0]?.webkitRelativePath.split('/')[0];
    document.getElementById('folder-name').textContent = folderName || 'No folder selected';
});

async function uploadImages() {
    const folderInput = document.getElementById('folderInput');
    const files = folderInput.files;
    const progressBar = document.getElementById('upload-progress');
    const progressContainer = document.getElementById('progress-container');
    const progressText = document.getElementById('progress-text');

    if (files.length === 0) {
        alert('Please select a folder');
        return;
    }

    progressContainer.style.display = 'block';
    const formData = new FormData();
    
    // Group files by their directories
    const filesByDir = {};
    for (let file of files) {
        const path = file.webkitRelativePath;
        const parts = path.split('/');
        if (parts.length >= 2) {
            const letter = parts[0].toUpperCase();
            if (!filesByDir[letter]) {
                filesByDir[letter] = [];
            }
            filesByDir[letter].push(file);
        }
    }

    // Add files to formData
    for (let letter in filesByDir) {
        filesByDir[letter].forEach(file => {
            formData.append('files', file);
            formData.append('letters', letter);
        });
    }

    try {
        const response = await fetch('/api/upload', {
            method: 'POST',
            body: formData
        });

        const results = await response.json();
        progressContainer.style.display = 'none';
        displayResults(results);
    } catch (error) {
        console.error('Error:', error);
        alert('An error occurred while uploading images');
        progressContainer.style.display = 'none';
    }
}

function displayResults(results) {
    const resultsDiv = document.getElementById('results');
    resultsDiv.innerHTML = '';
    
    results.forEach(item => {
        const resultCard = document.createElement('div');
        resultCard.className = 'result-item';
        resultCard.innerHTML = `
            <p><strong>Letter:</strong> ${item.letter}</p>
            <p><strong>Word:</strong> ${item.word}</p>
            <p><strong>Image:</strong><br>
                <a href="${item.imageUrl}" target="_blank">
                    View Image
                    <i class="fas fa-external-link-alt"></i>
                </a>
            </p>
        `;
        resultsDiv.appendChild(resultCard);
    });
}
