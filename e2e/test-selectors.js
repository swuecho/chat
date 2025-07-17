const { chromium } = require('playwright');

async function testSelectors() {
  const browser = await chromium.launch({ headless: true });
  const page = await browser.newPage();
  
  try {
    // Navigate to the app
    await page.goto('http://localhost:3002', { waitUntil: 'networkidle' });
    
    // Check if basic elements exist
    console.log('Checking if chat elements exist...');
    
    // Check if the main wrapper exists
    const imageWrapper = await page.$('#image-wrapper');
    console.log('Image wrapper exists:', !!imageWrapper);
    
    // Check if message textarea exists
    const messageTextarea = await page.$('#message_textarea textarea');
    console.log('Message textarea exists:', !!messageTextarea);
    
    // Check if we can find chat message structure
    const chatMessages = await page.$$('.chat-message');
    console.log('Number of chat messages found:', chatMessages.length);
    
    // Check if message-text class is accessible
    const messageTexts = await page.$$('.message-text');
    console.log('Number of message-text elements found:', messageTexts.length);
    
    // Check if regenerate buttons are accessible
    const regenerateButtons = await page.$$('.chat-message-regenerate');
    console.log('Number of regenerate buttons found:', regenerateButtons.length);
    
    console.log('✅ Basic selector test completed');
    
  } catch (error) {
    console.error('❌ Error during selector test:', error);
  }
  
  await browser.close();
}

// Check if the server is running
async function checkServer() {
  try {
    const response = await fetch('http://localhost:3002');
    return response.ok;
  } catch (error) {
    return false;
  }
}

async function main() {
  console.log('Testing message selectors after layout changes...');
  
  const serverRunning = await checkServer();
  if (!serverRunning) {
    console.log('⚠️  Server not running at http://localhost:3002');
    console.log('Please start the development server first: cd web && npm run dev');
    return;
  }
  
  await testSelectors();
}

main();