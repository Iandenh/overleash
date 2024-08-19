document.addEventListener('DOMContentLoaded',function () {
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
    const helpDialog = document.querySelector("#help-dialog");

    /**
     * @type {HTMLButtonElement}
     */
    const closeButtonHelpDialog = document.querySelector("dialog button");

    /**
     * @type {boolean}
     */
    let altMode = false

    const searchListener = () => {
        moveTo(-1, false);
    };

    searchBar.addEventListener('click', searchListener);

    closeButtonHelpDialog.addEventListener('click', () => {
        helpDialog.close()
    });

    const load = () => {
        currentIdx = -1;
        elements = document.querySelectorAll('.flag');
        searchBar = document.querySelector('.input');

        searchBar.removeEventListener('click', searchListener);
        searchBar.addEventListener('click', searchListener);

        const elementLength = elements.length;
        for (let i = 0; i < elementLength; i++) {
            elements[i].addEventListener('click', () => {
                if (currentIdx  === i) {
                    return;
                }

                moveTo(i, false);
            });
        }
    };

    load()

    document.addEventListener("keydown", (event) => {
        switch (event.key) {
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
                    refreshFlags();
                }
                return;
            case 'p':
                if (altMode) {
                    pauseOverrides();
                }
                return;
            case 'h':
                if (altMode) {
                    toggleHelp();
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
            case 'Alt':
                altMode = true;
                return;
        }
    });

    document.addEventListener("keyup", (event) => {
        switch (event.key) {
            case 'Alt':
                altMode = false;
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

    const refreshFlags = () => {
        htmx.trigger(".sync-btn", "refresh");
    }

    const pauseOverrides = () => {
        htmx.trigger(".pause-btn", "toggle-pause");
    }

    const toggleHelp = () => {
        if (helpDialog.open) {
            helpDialog.close()
        } else {
            helpDialog.showModal()
        }
    }
});
