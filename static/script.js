// static/script.js

// Function to fetch all stages from the API
async function fetchStages() {
    try {
        const response = await fetch('/stages');
        const stages = await response.json();
        return stages;
    } catch (error) {
        console.error('Error fetching stages:', error);
        return [];
    }
}

// Function to render stages on the page
async function renderStages() {
    const stagesContainer = document.getElementById('stages-container');
    stagesContainer.innerHTML = ''; // Clear previous content

    const stages = await fetchStages();
    stages.forEach(stage => {
        const stageElement = document.createElement('div');
        stageElement.classList.add('stage');

        const stageName = document.createElement('h2');
        stageName.textContent = stage.stage_name;
        stageElement.appendChild(stageName);

        const stagesList = document.createElement('ul');
        for (const [key, value] of Object.entries(stage.stages)) {
            const stageUrl = document.createElement('li');
            stageUrl.textContent = `${key}: ${value}`;
            stagesList.appendChild(stageUrl);
        }
        stageElement.appendChild(stagesList);

        const deleteButton = document.createElement('button');
        deleteButton.textContent = 'Delete';
        deleteButton.addEventListener('click', () => deleteStage(stage.id));
        stageElement.appendChild(deleteButton);

        stagesContainer.appendChild(stageElement);
    });
}

// Function to create a new stage via POST request
async function createStage(stageData) {
    try {
        const response = await fetch('/stages', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(stageData)
        });
        if (response.ok) {
            console.log('Stage created successfully!');
            renderStages(); // Refresh the stage list
        } else {
            console.error('Failed to create stage:', response.statusText);
        }
    } catch (error) {
        console.error('Error creating stage:', error);
    }
}

// Function to handle form submission and create a new stage
document.getElementById('create-stage-form').addEventListener('submit', async function(event) {
    event.preventDefault();
    const stageNameInput = document.getElementById('stage-name');
    const stageName = stageNameInput.value.trim();
    if (stageName) {
        const stageUrls = {};
        const inputs = Array.from(document.querySelectorAll('#stage-urls input'));
        for (let i = 0; i < inputs.length; i += 2) {
            const label = inputs[i].value.trim();
            const url = inputs[i + 1].value.trim();
            if (label && url) {
                stageUrls[label] = url;
            }
        }
        createStage({ stage_name: stageName, stages: stageUrls });
        stageNameInput.value = ''; // Clear the input field
        document.getElementById('stage-urls').innerHTML = ''; // Clear the URLs
        addUrlInputs(); // Add one empty set of URL inputs after submission
    } else {
        console.error('Stage name is required!');
    }
});

// Function to delete a stage
async function deleteStage(id) {
    try {
        const response = await fetch(`/stages/${id}`, {
            method: 'DELETE'
        });
        if (response.ok) {
            console.log('Stage deleted successfully!');
            renderStages(); // Refresh the stage list
        } else {
            console.error('Failed to delete stage:', response.statusText);
        }
    } catch (error) {
        console.error('Error deleting stage:', error);
    }
}

// Function to add input fields for stage URLs
function addUrlInputs() {
    const stageUrlsUl = document.getElementById('stage-urls');
    const newUrlInputs = document.createElement('li');
    
    const labelInput = document.createElement('input');
    labelInput.type = 'text';
    labelInput.name = `label${stageUrlsUl.childElementCount / 2 + 1}`;
    labelInput.placeholder = `Label ${stageUrlsUl.childElementCount / 2 + 1}`;
    
    const urlInput = document.createElement('input');
    urlInput.type = 'text';
    urlInput.name = `url${stageUrlsUl.childElementCount / 2 + 1}`;
    urlInput.placeholder = `URL ${stageUrlsUl.childElementCount / 2 + 1}`;
    
    newUrlInputs.appendChild(labelInput);
    newUrlInputs.appendChild(urlInput);
    
    stageUrlsUl.appendChild(newUrlInputs);
}


// Render stages when the page loads
window.onload = () => {
    renderStages();
    addUrlInputs();
};
