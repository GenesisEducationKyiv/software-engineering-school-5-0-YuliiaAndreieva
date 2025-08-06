export interface MailHogMessage {
  ID: string;
  From: {
    Mailbox: string;
    Domain: string;
  };
  To: Array<{
    Mailbox: string;
    Domain: string;
  }>;
  Content: {
    Headers: Record<string, string[]>;
    Body: string;
  };
}

export interface MailHogResponse {
  items: MailHogMessage[];
}

export async function getMailHogMessages(): Promise<MailHogMessage[]> {
  try {
    const response = await fetch('http://localhost:8025/api/v2/messages');
    if (!response.ok) {
      throw new Error(`Failed to fetch messages: ${response.status}`);
    }
    const data: MailHogResponse = await response.json();
    return data.items || [];
  } catch (error) {
    console.error('Error fetching MailHog messages:', error);
    return [];
  }
}

export async function clearMailHogMessages(): Promise<void> {
  try {
    const response = await fetch('http://localhost:8025/api/v1/messages', { 
      method: 'DELETE' 
    });
    if (!response.ok) {
      console.warn(`Failed to clear messages: ${response.status}`);
    }
  } catch (error) {
    console.error('Error clearing MailHog messages:', error);
  }
}

export async function waitForEmail(
  expectedEmail: string, 
  timeout = 10000
): Promise<MailHogMessage> {
  const startTime = Date.now();
  
  while (Date.now() - startTime < timeout) {
    const messages = await getMailHogMessages();
    const emailMessage = messages.find((msg) => 
      `${msg.To[0]?.Mailbox}@${msg.To[0]?.Domain}` === expectedEmail
    );
    
    if (emailMessage) {
      return emailMessage;
    }
    
    await new Promise(resolve => setTimeout(resolve, 500));
  }
  
  throw new Error(`Email to ${expectedEmail} not found within ${timeout}ms`);
}

function decodeEmailContent(content: string): string {
  let decoded = content.replace(/=3D/g, '=');
  decoded = decoded.replace(/=\s*\n/g, '\n');
  return decoded;
}

export function extractConfirmationLink(emailBody: string): string | null {
  const decodedBody = decodeEmailContent(emailBody);
  
  const cleanBody = decodedBody.replace(/\s+/g, ' ');
  
  const linkMatch = cleanBody.match(/href\s*=\s*"([^"]*\/api\/confirm\/[^"]*)"/);
  
  if (linkMatch) {
    const link = linkMatch[1].replace(/\s/g, '');
    return link;
  }
  
  return null;
}

export function debugEmailContent(emailMessage: MailHogMessage): void {
  console.log('=== Email Debug Info ===');
  console.log('From:', `${emailMessage.From.Mailbox}@${emailMessage.From.Domain}`);
  console.log('To:', `${emailMessage.To[0]?.Mailbox}@${emailMessage.To[0]?.Domain}`);
  console.log('Subject:', emailMessage.Content.Headers['Subject']);
  console.log('Original Body:', emailMessage.Content.Body);
  
  const decodedBody = decodeEmailContent(emailMessage.Content.Body);
  console.log('Decoded Body:', decodedBody);
  
  const cleanBody = decodedBody.replace(/\s+/g, ' ');
  console.log('Clean Body:', cleanBody);
  
  const linkMatch1 = cleanBody.match(/href\s*=\s*"([^"]*\/api\/confirm\/[^"]*)"/);
  console.log('Regex Match (with spaces):', linkMatch1);
  
  const anyLink = cleanBody.match(/href="([^"]*)"/);
  console.log('Any link found:', anyLink);
  
  const extractedLink = extractConfirmationLink(emailMessage.Content.Body);
  console.log('Extracted and cleaned link:', extractedLink);
  
  if (extractedLink) {
    const token = extractedLink.split('/api/confirm/')[1];
    console.log('Token from link:', token);
  }
  
  console.log('=== End Debug Info ===');
}
