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

async function reformatDogApp(req: NextApiRequest, res: NextApiResponse) {
  // Run the middleware
  await runMiddleware(req, res, cors)
  
  let response;
  console.log(req.body)
  // try {
  //   response = await sendgrid.send({
  //     to: 'tnelson@frewdev.com',
  //     from: 'nick@clyde.tech',
  //     subject: 'New message from Westdale Website',
  //     html: `${req.body.name} sent a message:
  //     <br /><br />
  //     ${req.body.message}
  //     <br /><br />
  //     Email address provided: <a href="mailto:${req.body.email}" target="_blank">${req.body.email}</a><br />
  //     Phone number provided: ${req.body.phone}`,
  //   });
  // } catch (error: any) {
  //   return res.status(error.statusCode || 500).json({ error: error.message });
  // }

  return res.status(200).json("logged");
}

export default reformatDogApp;