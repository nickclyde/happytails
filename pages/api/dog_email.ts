import type { NextApiRequest, NextApiResponse } from 'next';
import sendgrid from '@sendgrid/mail';
import Cors from 'cors'

sendgrid.setApiKey(process.env.SENDGRID_API_KEY || '');

const cors = Cors({
  origin: ['https://happytails.org', 'https://www.happytails.org'],
  methods: ['POST', 'GET', 'HEAD'],
})

// Helper method to wait for a middleware to execute before continuing
// And to throw an error when an error happens in a middleware
function runMiddleware(
  req: NextApiRequest,
  res: NextApiResponse,
  fn: Function
) {
  return new Promise((resolve, reject) => {
    fn(req, res, (result: any) => {
      if (result instanceof Error) {
        return reject(result)
      }

      return resolve(result)
    })
  })
}

function capitalizeFirstLetter(string: string) {
    return string.charAt(0).toUpperCase() + string.slice(1);
}

async function reformatDogApp(req: NextApiRequest, res: NextApiResponse) {
  // Run the middleware
  await runMiddleware(req, res, cors)
  
  let response;
  const data = req.body;
  const formattedData = Object.entries(data)
    .map(([key, value]) => {
      if (value == 'on') {
        value = 'Yes';
      }
      return `<b>${capitalizeFirstLetter(key).replace(/-/g, ' ').replace(/_/g, ' ')}:</b><br />${value}<br />`
    })
    .join('<br />');

  try {
    response = await sendgrid.send({
      to: 'dogs@happytails.org',
      from: 'nick@clyde.tech',
      replyTo: data.Email,
      subject: `${data['Name-of-dog-interested-in-adopting']} - Adoption Application from ${data.name}`,
      html: `${data.name} applied to adopt ${data['Name-of-dog-interested-in-adopting']}:
      <br /><br />
      ${formattedData}`,
    });
  } catch (error: any) {
    return res.status(error.statusCode || 500).json({ error: error.message });
  }

  return res.status(200).json(response);
}

export default reformatDogApp;