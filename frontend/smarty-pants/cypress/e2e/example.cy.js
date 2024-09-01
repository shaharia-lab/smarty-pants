// cypress/e2e/example.cy.js
describe('Home Page', () => {
    it('successfully loads with correct title', () => {
        cy.visit('/login', { timeout: 10000 }) // Increase timeout for slow connections
        cy.title().should('eq', 'SmartyPants AI') // Check the full title
    })
})