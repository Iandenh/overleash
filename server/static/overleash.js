const overleash = function() {
    "use strict";

    /**
     * @type {number}
     */
    let currentIdx = 0;
    /**
     * @type {NodeListOf<Element>}
     */
    let elements = [];

    /**
     * @type {HTMLTextAreaElement}
     */
    let searchBar = document.querySelector('.input');

    /**
     * @type {HTMLDialogElement}
     */
    let helpDialog = document.querySelector("#help-dialog");

    /**
     * @type {HTMLButtonElement}
     */
    let helpDialogBtn = document.querySelector(".search-btn");

    /**
     * @type {HTMLButtonElement}
     */
    let closeButtonHelpDialog = document.querySelector("dialog button");

    /**
     * @type {NodeListOf<HTMLButtonElement>}
     */
    let themeToggles = document.querySelectorAll(".theme-btn");

    const searchListener = () => {
        moveTo(-1, false);
    };
    const closeHelpDialogListener = () => {
        helpDialog.close();
    };

    /**
     * @param event {MouseEvent}
     */
    const outsideDialogClickListener = event => {
        console.log(event);
        if (event.target === event.currentTarget) {
            event.currentTarget.close()
        }
    }

    const load = () => {
        currentIdx = -1;
        elements = document.querySelectorAll('.flag');
        searchBar = document.querySelector('.input');
        closeButtonHelpDialog = document.querySelector("dialog button");
        helpDialog = document.querySelector("#help-dialog");
        themeToggles = document.querySelectorAll(".theme-btn");
        helpDialogBtn = document.querySelector(".search-btn");

        searchBar.removeEventListener('click', searchListener);
        searchBar.addEventListener('click', searchListener);
        closeButtonHelpDialog.removeEventListener('click', closeHelpDialogListener);
        closeButtonHelpDialog.addEventListener('click', closeHelpDialogListener);
        themeToggles.forEach(themeToggle => {
            themeToggle.removeEventListener('click', toggleTheme);
            themeToggle.addEventListener('click', toggleTheme);
        });
        helpDialogBtn.removeEventListener('click', toggleHelp);
        helpDialogBtn.addEventListener('click', toggleHelp);
        helpDialog.removeEventListener('mousedown', outsideDialogClickListener);
        helpDialog.addEventListener('mousedown', outsideDialogClickListener);

        const elementLength = elements.length;
        for (let i = 0; i < elementLength; i++) {
            elements[i].addEventListener('click', () => {
                if (currentIdx === i) {
                    return;
                }

                moveTo(i, false);
            });
        }
    };

    document.addEventListener("keydown", (event) => {
        const altMode = event.getModifierState('Alt');
        switch (event.key) {
            case 'Escape':
                if (helpDialog.open) {
                    toggleHelp(event)
                    return;
                }

                clearInput(event);
                return
            case 'ArrowDown':
                moveDown(event);
                return
            case 'ArrowUp':
                moveUp(event);
                return;
            case 'e':
                enable();
                return;
            case 'r':
                if (altMode) {
                    refreshFlags(event);
                }
                return;
            case 'p':
                if (altMode) {
                    pauseOverrides(event);
                }
                return;
            case 'h':
                if (altMode) {
                    toggleHelp(event);
                }
                return;
            case '?':
                if (document.activeElement !== searchBar) {
                    toggleHelp(event);
                }
                return;
            case 't':
                if (altMode) {
                    toggleTheme(event);
                }
                return;
            case 'd':
                disable();
                return;
            case 'q':
                remove();
                return;
            case 'i':
                toggleInfo();
                return;
            case '/':
                focusInput(event);
                return;
        }
    });

    htmx.on("htmx:afterSwap", function (event) {
        if (event.target.id === 'flags' || event.target === document.body) {
            load();
        }
    })

    /**
     * @param to {number}
     * @param focus {boolean}
     */
    const moveTo = (to, focus = true) => {
        elements[currentIdx]?.classList.remove('selected');

        currentIdx = to;

        if (!focus) {
            return;
        }

        elements[currentIdx]?.classList.add('selected');

        elements[currentIdx]?.scrollIntoView({
            behavior: 'auto',
            block: 'center',
            inline: 'center'
        });
    }

    /**
     * @param event {KeyboardEvent}
     */
    const moveDown = event => {
        if (currentIdx >= elements.length - 1) {
            return;
        }

        event.preventDefault();

        searchBar.blur();

        moveTo(currentIdx + 1);
    };

    /**
     * @param event {KeyboardEvent}
     */
    const moveUp = event => {
        if (currentIdx === -1) {
            return;
        }

        event.preventDefault();

        if (currentIdx === 0) {
            focus();
        }

        moveTo(currentIdx - 1);
    };

    const enable = () => {
        // Not in an element
        if (currentIdx === -1) {
            return;
        }
        htmx.trigger(elements[currentIdx], "enable-flag");
    };

    const toggleInfo = () => {
        // Not in an element
        if (currentIdx === -1) {
            return;
        }
        htmx.trigger(elements[currentIdx], "toggle-detail");
    };

    const disable = () => {
        // Not in an element
        if (currentIdx === -1) {
            return;
        }
        htmx.trigger(elements[currentIdx], "disable-flag");
    };

    const remove = () => {
        // Not in an element
        if (currentIdx === -1) {
            return;
        }
        htmx.trigger(elements[currentIdx], "remove-flag");
    };

    /**
     * @param event {KeyboardEvent}
     */
    const focusInput = (event) => {
        // Not in an element
        if (currentIdx === -1) {
            return;
        }

        event.preventDefault();

        elements[currentIdx]?.classList.remove('selected');

        currentIdx = -1;
        // Without this the key mapped to the focus is entered in the endpoint
        setTimeout(() => {
            focus();
        }, 0);
    };

    const focus = () => {
        searchBar.focus();

        // Apparently we need to this to let make it work
        setTimeout(() => {
            searchBar.setSelectionRange(searchBar.value.length, searchBar.value.length);
        }, 0);
    };

    /**
     * @param event {KeyboardEvent}
     */
    const refreshFlags = event => {
        event.preventDefault();

        htmx.trigger(".sync-btn", "refresh");
    }

    /**
     * @param event {KeyboardEvent}
     */
    const pauseOverrides = event => {
        event.preventDefault();

        htmx.trigger(".pause-btn", "toggle-pause");
    }

    /**
     * @param event {KeyboardEvent}
     */
    const clearInput = event => {
        if (currentIdx !== -1) {
            return;
        }

        event.preventDefault();

        searchBar.value = "";
        htmx.trigger(searchBar, "search");
    }

    /**
     * @param event {KeyboardEvent}
     */
    const toggleHelp = event => {
        event.preventDefault();
        if (helpDialog.open) {
            helpDialog.close();
        } else {
            helpDialog.showModal();
        }
    }

    /**
     * @param event {KeyboardEvent}
     */
    const toggleTheme = (event) => {
        event.preventDefault();

        let nextTheme = 'dark';
        const theme = document.documentElement.getAttribute('data-theme');
        if (theme === 'dark') {
            nextTheme = 'light';
        } else if (theme === 'light') {
            nextTheme = 'auto';
        }
        document.documentElement.setAttribute('data-theme', nextTheme);
        document.cookie = `prefers-color-scheme=${nextTheme};path=/;max-age=31536000;`;
    }

    load();
    focus();
}

if (document.readyState !== 'loading') {
    overleash();
} else {
    document.addEventListener('DOMContentLoaded', overleash);
}