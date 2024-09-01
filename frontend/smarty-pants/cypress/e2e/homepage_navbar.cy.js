describe('Homepage Navbar', () => {
    beforeEach(() => {
        cy.login() // Use the updated custom login command
    })

    it('displays the navbar with correct elements', () => {
        // No need to visit '/' again as login command already does this

        // Check for the logo and title
        cy.get('nav').find('svg').should('exist') // Logo
        cy.get('nav').contains('SmartyPants').should('be.visible')

        // Check for main navigation items
        const navItems = ['Home', 'Assistant', 'Datasources', 'AI Providers', 'Management']
        navItems.forEach(item => {
            cy.get('nav').contains(item).should('be.visible')
        })

        // Check for the logout button
        cy.get('nav').contains('Logout').should('be.visible')
    })

    /*it('navigates to correct pages when navbar items are clicked', () => {
        // Test navigation for Home
        cy.get('[data-testid="nav-home"]').click()
        cy.url().should('eq', Cypress.config().baseUrl + '/')

        // Test dropdown for Assistant
        cy.contains('Assistant').click()
        cy.contains('Conversation').click()
        cy.url().should('include', '/ask')

        // Test dropdown for Datasources
        cy.contains('Datasources').click()
        cy.contains('Documents').click()
        cy.url().should('include', '/documents')

        // Test dropdown for AI Providers
        cy.contains('AI Providers').click()
        cy.contains('LLM Providers').click()
        cy.url().should('include', '/llm-providers')

        // Test dropdown for Management
        cy.contains('Management').click()
        cy.contains('Settings').click()
        cy.url().should('include', '/settings')
    })*/

    it('logs out successfully', () => {
        cy.contains('Logout').click()
        cy.url().should('include', '/login')
    })
})