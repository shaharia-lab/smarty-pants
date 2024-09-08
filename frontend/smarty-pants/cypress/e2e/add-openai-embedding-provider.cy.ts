describe('Configure OpenAI Embedding Provider', () => {
    beforeEach(() => {
        cy.login();
        cy.visit('/embedding-providers');
    });

    it('should configure a new OpenAI embedding provider', () => {
        // Click on the "Configure" button for OpenAI provider
        cy.contains('OpenAI').parent().contains('button', 'Configure').click();

        // Verify we're on the configuration page
        cy.url().should('include', '/embedding-providers/openai');

        // Fill out the form
        cy.get('#name').type('Test OpenAI Provider');
        cy.get('#apiKey').type('test-api-key');
        cy.get('#modelId').select('text-embedding-ada-002');

        // Click the Validate button
        cy.get('button').contains('Validate').click();

        // Submit the form
        cy.get('button[type="submit"]').contains('Save Provider').click();

        // Assert that we're redirected back to the embedding providers page
        cy.url().should('include', '/embedding-providers');

        // Assert that a success message is displayed
        cy.contains('Embedding provider added successfully').should('be.visible');

        // Assert that the new provider appears in the list
        cy.contains('Test OpenAI Provider').should('be.visible');
    });

    it('should display an error message for invalid input', () => {
        // Click on the "Configure" button for OpenAI provider
        cy.contains('OpenAI').parent().contains('button', 'Configure').click();

        // Submit the form without filling it out
        cy.get('button[type="submit"]').contains('Save Provider').click();

        // Assert that error messages are displayed
        cy.contains('Name is required').should('be.visible');
        cy.contains('API Key is required').should('be.visible');
    });

    it('should allow editing an existing OpenAI provider', () => {
        // Assuming there's at least one configured OpenAI provider in the list
        cy.contains('Test OpenAI Provider').parent().contains('button', 'Edit').click();

        // Verify we're on the edit page
        cy.url().should('include', '/embedding-providers/openai/edit');

        // Update the name
        cy.get('#name').clear().type('Updated OpenAI Provider');

        // Submit the form
        cy.get('button[type="submit"]').contains('Update Provider').click();

        // Assert that we're redirected back to the embedding providers page
        cy.url().should('include', '/embedding-providers');

        // Assert that a success message is displayed
        cy.contains('Embedding provider updated successfully').should('be.visible');

        // Assert that the updated provider appears in the list
        cy.contains('Updated OpenAI Provider').should('be.visible');
    });
});