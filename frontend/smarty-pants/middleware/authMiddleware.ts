import { NextApiRequest, NextApiResponse } from 'next';
import authService from '@/services/authService';

export function authMiddleware(handler: (req: NextApiRequest, res: NextApiResponse) => Promise<void>) {
    return async (req: NextApiRequest, res: NextApiResponse) => {
        try {
            const authHeader = req.headers.authorization;

            if (!authHeader || !authHeader.startsWith('Bearer ')) {
                return res.status(401).json({ error: 'Missing or invalid Authorization header' });
            }

            const token = authHeader.split(' ')[1];

            if (!token) {
                return res.status(401).json({ error: 'Missing token' });
            }

            // Verify the token
            const isValid = await authService.verifyToken(token);

            if (!isValid) {
                return res.status(401).json({ error: 'Invalid or expired token' });
            }

            // Attach the verified token to the request for use in the handler
            (req as any).token = token;

            return handler(req, res);
        } catch (error) {
            console.error('Auth Middleware Error:', error);
            return res.status(500).json({ error: 'Internal Server Error' });
        }
    };
}