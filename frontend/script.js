document.addEventListener('DOMContentLoaded', () => {
    const API_BASE_URL = 'http://localhost:8080'; // Your Go API base URL

    // --- Element Selectors ---
    // All Games
    const loadAllGamesBtn = document.getElementById('loadAllGamesBtn');
    const allGamesOutput = document.getElementById('allGamesOutput');

    // Game by ID
    const gameIdInput = document.getElementById('gameIdInput');
    const getGameByIdBtn = document.getElementById('getGameByIdBtn');
    const deleteGameByIdBtn = document.getElementById('deleteGameByIdBtn');
    const gameByIdOutput = document.getElementById('gameByIdOutput');

    // Upload Log
    const logFileInput = document.getElementById('logFileInput');
    const uploadLogBtn = document.getElementById('uploadLogBtn');
    const uploadOutput = document.getElementById('uploadOutput');

    // Delete All Games
    const deleteAllGamesBtn = document.getElementById('deleteAllGamesBtn');
    const deleteAllOutput = document.getElementById('deleteAllOutput');

    // Global Player Ranking
    const loadPlayerRankingBtn = document.getElementById('loadPlayerRankingBtn');
    const playerRankingOutput = document.getElementById('playerRankingOutput');

    // --- Helper to display messages/data ---
    function displayData(element, data, isError = false) {
        if (typeof data === 'object') {
            element.textContent = JSON.stringify(data, null, 2);
        } else {
            element.textContent = data;
        }
        element.style.color = isError ? 'red' : 'green';
    }

    // --- API Call Functions ---

    // GET /games - Load All Games
    loadAllGamesBtn.addEventListener('click', async () => {
        allGamesOutput.textContent = 'Loading...';
        try {
            const response = await fetch(`${API_BASE_URL}/games`);
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({ error: `HTTP error! Status: ${response.status}` }));
                throw new Error(errorData.error || `HTTP error! Status: ${response.status}`);
            }
            const data = await response.json();
            displayData(allGamesOutput, data.length > 0 ? data : 'No games found.');
        } catch (error) {
            displayData(allGamesOutput, `Error: ${error.message}`, true);
        }
    });

    // GET /games/:id - Get Game by ID
    getGameByIdBtn.addEventListener('click', async () => {
        const gameId = gameIdInput.value.trim();
        if (!gameId) {
            displayData(gameByIdOutput, 'Please enter a Game ID.', true);
            return;
        }
        gameByIdOutput.textContent = 'Loading...';
        try {
            const response = await fetch(`${API_BASE_URL}/games/${gameId}`);
            const data = await response.json(); 
            if (!response.ok) {
                throw new Error(data.error || `HTTP error! Status: ${response.status}`);
            }
            displayData(gameByIdOutput, data);
        } catch (error) {
            displayData(gameByIdOutput, `Error: ${error.message}`, true);
        }
    });

    // DELETE /games/:id - Delete Game by ID
    deleteGameByIdBtn.addEventListener('click', async () => {
        const gameId = gameIdInput.value.trim();
        if (!gameId) {
            displayData(gameByIdOutput, 'Please enter a Game ID to delete.', true);
            return;
        }
        if (!confirm(`Are you sure you want to delete game ID: ${gameId}?`)) {
            return;
        }
        gameByIdOutput.textContent = 'Deleting...';
        try {
            const response = await fetch(`${API_BASE_URL}/games/${gameId}`, { method: 'DELETE' });
            const data = await response.json(); 
            if (!response.ok) {
                 throw new Error(data.error || `HTTP error! Status: ${response.status}`);
            }
            displayData(gameByIdOutput, data.message || 'Game deleted successfully.');
            loadAllGamesBtn.click(); // Refresh the list of all games
        } catch (error) {
            displayData(gameByIdOutput, `Error: ${error.message}`, true);
        }
    });

    // POST /games/upload - Upload Log File
    uploadLogBtn.addEventListener('click', async () => {
        const file = logFileInput.files[0];
        if (!file) {
            displayData(uploadOutput, 'Please select a log file to upload.', true);
            return;
        }
        uploadOutput.textContent = 'Uploading...';
        const formData = new FormData();
        formData.append('logFile', file);

        try {
            const response = await fetch(`${API_BASE_URL}/games/upload`, {
                method: 'POST',
                body: formData,
            });
            const data = await response.json(); 
            if (!response.ok) {
                throw new Error(data.error || `HTTP error! Status: ${response.status}`);
            }
            displayData(uploadOutput, data.message || 'File uploaded successfully.');
            loadAllGamesBtn.click(); // Refresh the list of all games
        } catch (error) {
            displayData(uploadOutput, `Error: ${error.message}`, true);
        }
    });

    // DELETE /games - Delete All Games
    deleteAllGamesBtn.addEventListener('click', async () => {
        if (!confirm('Are you sure you want to delete ALL games? This action cannot be undone.')) {
            return;
        }
        deleteAllOutput.textContent = 'Deleting all...';
        try {
            const response = await fetch(`${API_BASE_URL}/games`, { method: 'DELETE' });
            const data = await response.json();
            if (!response.ok) {
                throw new Error(data.error || `HTTP error! Status: ${response.status}`);
            }
            displayData(deleteAllOutput, data.message || 'All games deleted successfully.');
            loadAllGamesBtn.click(); // Refresh the list of all games
        } catch (error) {
            displayData(deleteAllOutput, `Error: ${error.message}`, true);
        }
    });

    // GET /playersranking - Load Global Player Ranking
    loadPlayerRankingBtn.addEventListener('click', async () => {
        playerRankingOutput.textContent = 'Loading ranking...'; // Keep this for initial feedback
        playerRankingOutput.style.color = '#333'; // Reset color
        try {
            const response = await fetch(`${API_BASE_URL}/playersranking`);
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({ error: `HTTP error! Status: ${response.status}` }));
                throw new Error(errorData.error || `HTTP error! Status: ${response.status}`);
            }
            const data = await response.json();
            if (data && data.length > 0) {
                let rankingHtml = '<h3>Player Rankings:</h3><ul>';
                data.forEach(player => {
                    rankingHtml += `<li>${player.player_name}: ${player.total_kills} kills</li>`;
                });
                rankingHtml += '</ul>';
                playerRankingOutput.innerHTML = rankingHtml; // Use innerHTML to render HTML
            } else {
                playerRankingOutput.textContent = 'No player ranking data found.';
            }
        } catch (error) {
            // Use the existing displayData for errors, or customize error display here too
            displayData(playerRankingOutput, `Error: ${error.message}`, true);
        }
    });
}); 