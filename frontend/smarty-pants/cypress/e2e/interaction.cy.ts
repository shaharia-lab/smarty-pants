describe('Chat Interface Initial Load', () => {
    beforeEach(() => {
        // Intercept API calls
        cy.intercept('GET', '**/api/v1/interactions').as('getChatHistories')
        cy.intercept('POST', '**/api/v1/interactions').as('startNewSession')

        // Login and visit the chat page
        cy.login()
        cy.visit('http://localhost:3000/ask')
    })

    it('should load chat histories and start a new session', () => {
        // Verify that the chat histories API was called
        cy.wait('@getChatHistories').then((interception) => {
            expect(interception.response.statusCode).to.equal(200)
            // Log the response body for debugging
            cy.log('Chat histories response:', JSON.stringify(interception.response.body))
        })

        // Check if chat histories are displayed
        cy.get('h2').contains('Chat Histories').should('be.visible')

        // Check for either chat history items or the "No chat histories available" message
        cy.get('ul.divide-y li, p.text-gray-500').should('exist').then($elements => {
            if ($elements.length > 0) {
                if ($elements.is('li')) {
                    cy.wrap($elements).should('have.length.at.least', 1)
                } else {
                    cy.wrap($elements).should('contain', 'No chat histories available')
                }
            }
        })

        // Verify that a new session was started automatically
        cy.wait('@startNewSession').its('response.statusCode').should('eq', 200)

        // Check if the chat box appears
        cy.get('h2').contains('Chat Session').should('be.visible')

        // Verify an initial system message is displayed
        cy.get('.overflow-y-auto .bg-gray-100')
            .first()
            .should('exist')
            .and('not.be.empty')
            .then(($div) => {
                cy.log('Initial message:', $div.text())
            })

        // Check if the input area and send button are present
        cy.get('textarea[placeholder="Type your message here... (Shift+Enter for new line)"]').should('be.visible')
        cy.get('button').contains('Send').should('be.visible')

        // Log the entire page content for debugging
        cy.log('Page content:')
        cy.document().then((doc) => {
            cy.log(doc.body.innerHTML)
        })
    })

    it('should handle errors when starting a new session', () => {
        // Intercept and mock an error response for starting a new session
        cy.intercept('POST', '**/api/v1/interactions', {
            statusCode: 500,
            body: 'Server error'
        }).as('startNewSessionError')

        // Reload the page to trigger the error
        cy.reload()

        // Wait for the error response
        cy.wait('@startNewSessionError').then((interception) => {
            expect(interception.response.statusCode).to.equal(500)
            cy.log('Error response:', interception.response.body)
        })

        // Check that the application doesn't crash and remains in a usable state
        cy.get('h2').contains('Chat Histories').should('be.visible')

        // Check for either chat history items or the "No chat histories available" message
        cy.get('ul.divide-y li, p.text-gray-500').should('exist').then($elements => {
            if ($elements.length > 0) {
                if ($elements.is('li')) {
                    cy.wrap($elements).should('have.length.at.least', 1)
                } else {
                    cy.wrap($elements).should('contain', 'No chat histories available')
                }
            }
        })

        // Check if the chat interface is still present, but it might be in an error state
        cy.get('h2').contains('Chat Session').should('be.visible')
        cy.get('textarea').should('exist')
        cy.get('button').contains('Send').should('exist')

        // Verify that the "Start New Session" button is still clickable
        cy.contains('button', 'Start New Session')
            .should('be.visible')
            .and('not.be.disabled')

        // Check for any error messages displayed to the user
        cy.get('body').then($body => {
            if ($body.find('.error-message').length > 0) {
                cy.get('.error-message').should('be.visible')
            }
        })

        // Log the content of the chat area for debugging
        cy.get('.overflow-y-auto')
            .then(($chatArea) => {
                cy.log('Chat area content after error:', $chatArea.text())
            })

        // Log the entire page content for debugging
        cy.log('Page content after error:')
        cy.document().then((doc) => {
            cy.log(doc.body.innerHTML)
        })
    })

    it('should start a new session when the button is clicked', () => {
        // Click the "Start New Session" button
        cy.contains('button', 'Start New Session').click()

        // Verify that a new session API call is made
        cy.wait('@startNewSession').its('response.statusCode').should('eq', 200)

        // Check that the chat messages are cleared and a new initial message is present
        cy.get('.overflow-y-auto .bg-gray-100')
            .should('have.length', 1) // Only the initial message should be present
            .first()
            .should('exist')
            .and('not.be.empty')
            .then(($div) => {
                cy.log('New session initial message:', $div.text())
            })
    })
})