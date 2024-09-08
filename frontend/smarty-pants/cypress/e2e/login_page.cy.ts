// cypress/e2e/authentication_flow.cy.js
describe('Authentication Flow', () => {
    it('redirects unauthenticated user to login, completes OAuth flow, and accesses protected content', () => {
        // Intercept the initial redirect to login
        cy.intercept('GET', '/login').as('loginRedirect')

        // Visit the root URL
        cy.visit('/', { timeout: 10000 })

        // Assert that we've been redirected to the login page
        cy.url().should('include', '/login')

        // Check the page title
        cy.title().should('eq', 'SmartyPants AI')

        // Intercept the AJAX call
        cy.intercept('POST', '/api/v1/auth/initiate').as('initiateAuth')

        // Intercept the OAuth2 redirect
        cy.intercept('GET', '**/authorize**').as('oauthRedirect')

        // Intercept the final redirect back to your app
        cy.intercept('GET', '/login/callback**').as('loginCallback')

        // Check if the button exists and click it
        cy.contains('button', 'Sign in with Google')
            .should('exist')
            .and('be.visible')
            .click()

        // Wait for the AJAX call and check its response
        cy.wait('@initiateAuth').then((interception) => {
            expect(interception.response.statusCode).to.equal(200)
        })

        // Wait for the OAuth2 redirect
        cy.wait('@oauthRedirect').then((interception) => {
            expect(interception.request.url).to.include('/authorize')
        })

        // Wait for 3 seconds (simulating automatic approval)
        cy.wait(3000)

        // Assert we're back at the root URL (or wherever your app redirects after login)
        cy.url().should('eq', Cypress.config().baseUrl + '/')

        // Check for elements that indicate a logged-in state
        cy.get('nav').should('contain', 'Logout')

        // Optional: Check for the presence of protected content
        cy.get('body').should('contain', 'Dashboard') // Adjust based on your actual content
    })
})