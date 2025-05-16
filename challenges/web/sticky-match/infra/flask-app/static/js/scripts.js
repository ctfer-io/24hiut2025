// This file contains JavaScript code for the application to enhance interactivity.

document.addEventListener('DOMContentLoaded', function() {
    function refreshStats() {
        fetch('/stats')
            .then(res => res.json())
            .then(data => {
                document.getElementById('latency').innerText = (data.average_latency * 1000).toFixed(6) + ' ms';
            });
    }

    setInterval(refreshStats, 1000);
    refreshStats();
});