import type { NextApiRequest, NextApiResponse } from 'next';
import sendgrid from '@sendgrid/mail';

sendgrid.setApiKey(process.env.SENDGRID_API_KEY || '');

async function reformatDogApp(req: NextApiRequest, res: NextApiResponse) {
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