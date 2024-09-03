describe('Datasource Homepage', () => {
    beforeEach(() => {
        cy.login(); // Assuming this command exists and logs the user in
    });

    it('displays the correct layout with configured and available datasources', () => {
        cy.visit('/datasources');
        cy.get('h2').should('have.length', 2);
        cy.contains('h2', 'Configured Datasources').should('be.visible');
        cy.contains('h2', 'Available Datasources').should('be.visible');
    });

    context('With configured datasources', () => {
        beforeEach(() => {
            cy.intercept('GET', '**/api/v1/datasource', {
                statusCode: 200,
                body: {
                    datasources: [
                        { uuid: '1', name: 'Test Datasource 1', source_type: 'slack', status: 'active' },
                        { uuid: '2', name: 'Test Datasource 2', source_type: 'github', status: 'inactive' },
                    ],
                },
            }).as('getDatasources');
            cy.visit('/datasources');
            cy.wait('@getDatasources');
        });

        it('displays configured datasources with correct actions', () => {
            cy.contains('Test Datasource 1').should('be.visible');
            cy.contains('Test Datasource 2').should('be.visible');

            cy.contains('Test Datasource 1').closest('.border-b').within(() => {
                cy.get('a').contains('Edit').should('be.visible');
                cy.get('button').contains('Deactivate').should('be.visible');
                cy.get('button').contains('Delete').should('be.visible');
            });

            cy.contains('Test Datasource 2').closest('.border-b').within(() => {
                cy.get('a').contains('Edit').should('be.visible');
                cy.get('button').contains('Activate').should('be.visible');
                cy.get('button').contains('Delete').should('be.visible');
            });
        });
    });

    context('With no configured datasources', () => {
        beforeEach(() => {
            cy.intercept('GET', '**/api/v1/datasource', {
                statusCode: 200,
                body: {
                    datasources: [],
                },
            }).as('getEmptyDatasources');
            cy.visit('/datasources');
            cy.wait('@getEmptyDatasources');
        });

        it('displays a message when no datasources are configured', () => {
            cy.contains('Configured Datasources').parent().within(() => {
                cy.contains('No items to display.').should('be.visible');
            });
        });
    });

    it('displays available datasources', () => {
        cy.visit('/datasources');
        cy.contains('Available Datasources').parent().within(() => {
            cy.get('div[class*="border-b"]').should('have.length.at.least', 1);
            cy.get('div[class*="border-b"]').first().within(() => {
                cy.get('h3').should('be.visible');
                cy.get('p').should('be.visible');
                cy.contains('Configure').should('be.visible');
            });
        });
    });
});