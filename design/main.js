(function () {
    document
        .getElementById("search-form")
        .addEventListener("click", function (event) {
            const input = document.getElementById("search-input");
            if (event.target !== input && event.target.type !== "submit") {
                input.focus();
            }
        });
})();
