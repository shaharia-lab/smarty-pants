Cypress.Commands.add('login', () => {
    cy.intercept('GET', '/login').as('loginRedirect')
    cy.intercept('POST', '/api/v1/auth/initiate').as('initiateAuth')
    cy.intercept('GET', '**/authorize**').as('oauthRedirect')

    cy.visit('/')
    cy.url().should('include', '/login')

    cy.contains('button', 'Sign in with Google').click()

    cy.wait('@initiateAuth').its('response.statusCode').should('eq', 200)
    cy.wait('@oauthRedirect')

    // Instead of waiting for a fixed time, let's wait for the URL to change
    cy.url().should('eq', Cypress.config().baseUrl + '/')

    // Check for logged-in state
    cy.get('nav').should('contain', 'Logout')

    // Additional check to ensure the page has loaded completely
    cy.get('body').should('not.have.class', 'loading')
})

Cypress.Commands.add('logout', () => {
    // Click the logout button
    cy.get('nav').contains('Logout').click()

    // Wait for redirection to login page
    cy.url().should('include', '/login')

    // Verify that the login button is visible (indicating logged out state)
    cy.contains('button', 'Sign in with Google').should('be.visible')

    // Optional: Clear any local storage or cookies
    cy.clearLocalStorage()
    cy.clearCookies()
})