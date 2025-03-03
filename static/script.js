document.getElementById('folderInput').addEventListener('change', function(e) {
    const folderName = e.target.files[0]?.webkitRelativePath.split('/')[0];
    document.getElementById('folder-name').textContent = folderName || 'No folder selected';
});

async function uploadImages() {
    const folderInput = document.getElementById('folderInput');
    const processNested = document.getElementById('processNestedFolders').checked;
    const files = folderInput.files;
    const progressContainer = document.getElementById('progress-container');

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
        
        // Handle nested folders
        if (processNested) {
            // Process all folders that contain images
            const parentFolder = parts[parts.length - 2].toUpperCase();
            if (parts.length >= 2 && file.type.startsWith('image/')) {
                if (!filesByDir[parentFolder]) {
                    filesByDir[parentFolder] = [];
                }
                filesByDir[parentFolder].push(file);
            }
        } else {
            // Original behavior - only process first level folders
            if (parts.length === 2) {
                const letter = parts[0].toUpperCase();
                if (!filesByDir[letter]) {
                    filesByDir[letter] = [];
                }
                filesByDir[letter].push(file);
            }
        }
    }

    // Add files to formData with folder info
    for (let letter in filesByDir) {
        filesByDir[letter].forEach(file => {
            formData.append('files', file);
            formData.append('letters', letter);
            formData.append('paths', file.webkitRelativePath);
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
    
    results.forEach(letterGroup => {
        const letterSection = document.createElement('div');
        letterSection.className = 'letter-section';
        letterSection.innerHTML = `<h2>Letter ${letterGroup.letter}</h2>`;
        
        letterGroup.words.forEach(item => {
            const resultCard = document.createElement('div');
            resultCard.className = 'result-item';
            resultCard.innerHTML = `
                <p><strong>Word:</strong> ${item.word}</p>
                <p><strong>Image:</strong><br>
                    <a href="${item.imageUrl}" target="_blank">
                        View Image
                        <i class="fas fa-external-link-alt"></i>
                    </a>
                </p>
            `;
            letterSection.appendChild(resultCard);
        });
        
        resultsDiv.appendChild(letterSection);
    });
}
