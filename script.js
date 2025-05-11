const deleteAllGamesBtn = document.getElementById('deleteAllGamesBtn');
const deleteAllOutput = document.getElementById('deleteAllOutput');

// Global Player Ranking
const loadPlayerRankingBtn = document.getElementById('loadPlayerRankingBtn');
const playerRankingOutput = document.getElementById('playerRankingOutput');

// --- Helper to display messages/data ---
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

// GET /playersranking - Load Global Player Ranking
loadPlayerRankingBtn.addEventListener('click', async () => {
    playerRankingOutput.textContent = 'Loading ranking...';
    try {
        const response = await fetch(`${API_BASE_URL}/playersranking`);
        if (!response.ok) {
            const errorData = await response.json().catch(() => ({ error: `HTTP error! Status: ${response.status}` }));
            throw new Error(errorData.error || `HTTP error! Status: ${response.status}`);
        }
        const data = await response.json();
        if (data && data.length > 0) {
            // Format the ranking for better display
            let formattedRanking = "Player Rankings:\n------------------\n";
            data.forEach(player => {
                formattedRanking += `${player.name}: ${player.score} kills\n`;
            });
            displayData(playerRankingOutput, formattedRanking.trim());
        } else {
            displayData(playerRankingOutput, 'No player ranking data found.');
        }
    } catch (error) {
        displayData(playerRankingOutput, `Error: ${error.message}`, true);
    }
}); 